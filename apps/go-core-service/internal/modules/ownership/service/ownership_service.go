package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/publisher"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/rabbitmq"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/types"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/entity"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/repository"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	auditlog "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/audit_log"
	"gorm.io/gorm"
)

type IOwnershipService interface {
	CustomerRequestOTP(ctx context.Context, req dto.CustomerRequestOTPReq, userID uuid.UUID) (*uuid.UUID, error)
	CustomerVerifyAndRegister(ctx context.Context, req dto.CustomerVerifyAndRegisterReq, userID uuid.UUID) (*entity.Ownership, error)

	AdminRequestOTP(ctx context.Context, req dto.AdminRequestOTPReq, adminID uuid.UUID) (*uuid.UUID, error)
	AdminVerifyAndRegister(ctx context.Context, req dto.AdminVerifyAndRegisterReq, adminID uuid.UUID) (*entity.Ownership, error)

	ApproveOwnership(ctx context.Context, ownershipID uuid.UUID, adminID uuid.UUID) error
	RejectOwnership(ctx context.Context, ownershipID uuid.UUID, adminID uuid.UUID) error

	// UC-P1-OWNER-02: Xem thông tin chi tiết quyền sở hữu
	GetOwnershipDetail(ctx context.Context, productItemID uuid.UUID) (*dto.OwnershipDetailRes, error)
	
	// CRUD Extensions
	TransferOwnership(ctx context.Context, id uuid.UUID, req dto.TransferOwnershipReq, currentUserID uuid.UUID, role string) error
	DeleteOwnership(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, role string) error
	SearchOwnerships(ctx context.Context, req dto.SearchOwnershipsReq, currentUserID uuid.UUID, role string) (*dto.PaginatedOwnershipsRes, error)
}

type OwnershipService struct {
	repo         repository.IOwnershipRepository
	productPort  IProductService
	emailClient  IEmailOTPClient
	userProvider IUserInfoProvider
	pub          *publisher.Publisher
	auditLog     auditlog.AuditLogService
}

func NewOwnershipService(
	repo repository.IOwnershipRepository,
	productPort IProductService,
	emailClient IEmailOTPClient,
	userProvider IUserInfoProvider,
	pub *publisher.Publisher,
	auditLog auditlog.AuditLogService,
) IOwnershipService {
	return &OwnershipService{
		repo:         repo,
		productPort:  productPort,
		emailClient:  emailClient,
		userProvider: userProvider,
		pub:          pub,
		auditLog:     auditLog,
	}
}

// ---------------------------------------------------------------------------
// VALIDATION HELPERS
// ---------------------------------------------------------------------------

func (s *OwnershipService) ensureItemAvailableForRegistration(ctx context.Context, productItemID uuid.UUID) error {
	filterActive := repository.SearchFilter{
		ProductItemIDs:  []uuid.UUID{productItemID},
		OwnershipStatus: string(entity.OwnershipStatusActive),
		Limit:           1,
	}
	activeList, _, err := s.repo.SearchOwnerships(ctx, filterActive)
	if err == nil && len(activeList) > 0 {
		return apperror.NewValidation("Thiết bị này đã có người đăng ký sở hữu")
	}

	filterPending := repository.SearchFilter{
		ProductItemIDs:  []uuid.UUID{productItemID},
		OwnershipStatus: string(entity.OwnershipStatusPending),
		Limit:           1,
	}
	pendingList, _, err := s.repo.SearchOwnerships(ctx, filterPending)
	if err == nil && len(pendingList) > 0 {
		return apperror.NewValidation("Thiết bị này đang chờ duyệt đăng ký sở hữu")
	}

	return nil
}

// ---------------------------------------------------------------------------
// CUSTOMER FLOW
// ---------------------------------------------------------------------------

func (s *OwnershipService) CustomerRequestOTP(ctx context.Context, req dto.CustomerRequestOTPReq, userID uuid.UUID) (*uuid.UUID, error) {
	email, _, _, err := s.userProvider.GetUserEmailByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	productID, err := s.productPort.FindProductByQR(ctx, req.QRCode)
	if err != nil {
		return nil, err
	}

	if err := s.ensureItemAvailableForRegistration(ctx, productID); err != nil {
		return nil, err
	}

	if err := s.productPort.ValidateProductOwnershipStatus(ctx, productID); err != nil {
		return nil, err
	}

	err = s.emailClient.RequestOTP(ctx, email, productID.String())
	if err != nil {
		return nil, err
	}
	
	return &productID, nil
}

func (s *OwnershipService) CustomerVerifyAndRegister(ctx context.Context, req dto.CustomerVerifyAndRegisterReq, userID uuid.UUID) (*entity.Ownership, error) {
	email, _, _, err := s.userProvider.GetUserEmailByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	valid, err := s.emailClient.ValidateOTP(ctx, email, req.OTP)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, errors.New("invalid or expired OTP")
	}

	if err := s.ensureItemAvailableForRegistration(ctx, req.ProductID); err != nil {
		return nil, err
	}

	var saved *entity.Ownership
	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		txCtx := entity.WithTx(ctx, tx)

		newOwnership := entity.NewOwnership(req.ProductID, userID)
		newOwnership.Status = entity.OwnershipStatusPending // Customers start as pending

		saved, err = s.repo.CreateOwnership(txCtx, tx, newOwnership)
		if err != nil {
			return err
		}

		if err := s.productPort.UpdateOwnershipStatus(txCtx, req.ProductID, string(entity.OwnershipStatusPending)); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	_ = s.auditLog.LogCreate(ctx, &userID, "Ownership", saved.ID, saved)

	return saved, nil
}

// ---------------------------------------------------------------------------
// ADMIN FLOW
// ---------------------------------------------------------------------------

func (s *OwnershipService) AdminRequestOTP(ctx context.Context, req dto.AdminRequestOTPReq, adminID uuid.UUID) (*uuid.UUID, error) {
	productID, err := s.productPort.FindProductByQR(ctx, req.QRCode)
	if err != nil {
		return nil, err
	}

	if err := s.ensureItemAvailableForRegistration(ctx, productID); err != nil {
		return nil, err
	}

	if err := s.productPort.ValidateProductOwnershipStatus(ctx, productID); err != nil {
		return nil, err
	}

	// Verify user exists BEFORE sending OTP
	_, err = s.userProvider.EnsureUserExists(ctx, req.OwnerEmail, req.OwnerName, req.OwnerPhone)
	if err != nil {
		return nil, err
	}

	err = s.emailClient.RequestOTP(ctx, req.OwnerEmail, productID.String())
	if err != nil {
		return nil, err
	}
	
	return &productID, nil
}

func (s *OwnershipService) AdminVerifyAndRegister(ctx context.Context, req dto.AdminVerifyAndRegisterReq, adminID uuid.UUID) (*entity.Ownership, error) {
	valid, err := s.emailClient.ValidateOTP(ctx, req.OwnerEmail, req.OTP)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, errors.New("invalid or expired OTP")
	}

	if err := s.ensureItemAvailableForRegistration(ctx, req.ProductID); err != nil {
		return nil, err
	}

	ownerID, err := s.userProvider.EnsureUserExists(ctx, req.OwnerEmail, req.OwnerName, req.OwnerPhone)
	if err != nil {
		return nil, err
	}

	var saved *entity.Ownership
	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		txCtx := entity.WithTx(ctx, tx)

		newOwnership := entity.NewOwnership(req.ProductID, ownerID)
		newOwnership.Status = entity.OwnershipStatusActive // Admins register directly as Active

		saved, err = s.repo.CreateOwnership(txCtx, tx, newOwnership)
		if err != nil {
			return err
		}

		if err := s.productPort.UpdateOwnershipStatus(txCtx, req.ProductID, string(entity.OwnershipStatusActive)); err != nil {
			return err
		}

		// Insert Event to database
		if tx != nil {
			eventObj := map[string]interface{}{
				"id":              uuid.New(),
				"product_item_id": req.ProductID,
				"actor_id":        adminID,
				"event_type":      "REGISTERED",
				"title":           "Đăng ký sở hữu",
				"description":     "Admin đăng ký sở hữu trực tiếp cho khách hàng",
				"created_at":      time.Now(),
			}
			if err := tx.Table("events").Create(&eventObj).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	_ = s.auditLog.LogCreate(ctx, &adminID, "Ownership", saved.ID, saved)

	// Publish to RabbitMQ after commit
	if s.pub != nil {
		event := types.Event{
			EventID:       uuid.NewString(),
			EventType:     rabbitmq.OwnerCreatedRK,
			EventVersion:  "1.0",
			Timestamp:     time.Now().UTC(),
			Producer:      "go-core-service",
			CorrelationID: uuid.NewString(),
			Payload: map[string]interface{}{
				"ownership_id":    saved.ID.String(),
				"product_item_id": saved.ProductItemID.String(),
				"owner_id":        saved.OwnerID.String(),
				"status":          "ACTIVE",
				"registered_by":   adminID.String(),
				"registered_at":   time.Now().UTC(),
			},
		}
		_ = s.pub.Publish(event)
	}

	return saved, nil
}

// ---------------------------------------------------------------------------
// ADMIN APPROVAL (FR-037 Extension)
// ---------------------------------------------------------------------------

func (s *OwnershipService) ApproveOwnership(ctx context.Context, ownershipID uuid.UUID, adminID uuid.UUID) error {
	own, err := s.repo.GetOwnershipByID(ctx, ownershipID)
	if err != nil {
		return err
	}

	if own.Status != entity.OwnershipStatusPending {
		return errors.New("chỉ có thể duyệt quyền sở hữu đang ở trạng thái PENDING")
	}

	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		txCtx := entity.WithTx(ctx, tx)

		err = s.repo.UpdateOwnershipStatusAndEndedAt(txCtx, tx, ownershipID, string(entity.OwnershipStatusActive))
		if err != nil {
			return err
		}

		if err := s.productPort.UpdateOwnershipStatus(txCtx, own.ProductItemID, string(entity.OwnershipStatusActive)); err != nil {
			return err
		}

		// Insert Event to database
		if tx != nil {
			eventObj := map[string]interface{}{
				"id":              uuid.New(),
				"product_item_id": own.ProductItemID,
				"actor_id":        adminID,
				"event_type":      "REGISTERED",
				"title":           "Đăng ký sở hữu",
				"description":     "Duyệt đăng ký sở hữu cho chủ sở hữu mới",
				"created_at":      time.Now(),
			}
			if err := tx.Table("events").Create(&eventObj).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	if updatedOwn, err := s.repo.GetOwnershipByID(ctx, ownershipID); err == nil && updatedOwn != nil {
		_ = s.auditLog.LogUpdate(ctx, &adminID, "Ownership", own.ID, own, updatedOwn)
	}

	// Publish to RabbitMQ after commit
	if s.pub != nil {
		event := types.Event{
			EventID:       uuid.NewString(),
			EventType:     rabbitmq.OwnerCreatedRK,
			EventVersion:  "1.0",
			Timestamp:     time.Now().UTC(),
			Producer:      "go-core-service",
			CorrelationID: uuid.NewString(),
			Payload: map[string]interface{}{
				"ownership_id":    own.ID.String(),
				"product_item_id": own.ProductItemID.String(),
				"owner_id":        own.OwnerID.String(),
				"status":          "ACTIVE",
				"approved_by":     adminID.String(),
				"approved_at":     time.Now().UTC(),
			},
		}
		_ = s.pub.Publish(event)
	}

	return nil
}

func (s *OwnershipService) RejectOwnership(ctx context.Context, ownershipID uuid.UUID, adminID uuid.UUID) error {
	own, err := s.repo.GetOwnershipByID(ctx, ownershipID)
	if err != nil {
		return err
	}

	if own.Status != entity.OwnershipStatusPending {
		return errors.New("chỉ có thể từ chối quyền sở hữu đang ở trạng thái PENDING")
	}

	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		txCtx := entity.WithTx(ctx, tx)

		// Reject by setting status to REVOKED.
		err = s.repo.UpdateOwnershipStatusAndEndedAt(txCtx, tx, ownershipID, string(entity.OwnershipStatusRevoked))
		if err != nil {
			return err
		}

		if err := s.productPort.UpdateOwnershipStatus(txCtx, own.ProductItemID, "IN_STOCK"); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	if updatedOwn, err := s.repo.GetOwnershipByID(ctx, ownershipID); err == nil && updatedOwn != nil {
		_ = s.auditLog.LogUpdate(ctx, &adminID, "Ownership", own.ID, own, updatedOwn)
	}

	// Publish to RabbitMQ after commit
	if s.pub != nil {
		event := types.Event{
			EventID:       uuid.NewString(),
			EventType:     rabbitmq.OwnerDeletedRK,
			EventVersion:  "1.0",
			Timestamp:     time.Now().UTC(),
			Producer:      "go-core-service",
			CorrelationID: uuid.NewString(),
			Payload: map[string]interface{}{
				"ownership_id":    own.ID.String(),
				"product_item_id": own.ProductItemID.String(),
				"status":          "REVOKED",
				"rejected_by":     adminID.String(),
				"rejected_at":     time.Now().UTC(),
			},
		}
		_ = s.pub.Publish(event)
	}

	return nil
}

// ---------------------------------------------------------------------------
// DETAIL API (UC-P1-OWNER-02)
// ---------------------------------------------------------------------------

func (s *OwnershipService) GetOwnershipDetail(ctx context.Context, productItemID uuid.UUID) (*dto.OwnershipDetailRes, error) {
	// 1. Lấy record sở hữu hiện tại (mới nhất) — VAL-002: phân biệt "ownership not found"
	own, err := s.repo.GetOwnershipByProductItemID(ctx, productItemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound("ownership information")
		}
		return nil, err
	}

	// 2. Lấy lịch sử đăng ký quyền sở hữu (AC-002)
	history, err := s.repo.GetOwnershipHistoryByProductItemID(ctx, productItemID)
	if err != nil {
		// Không block — lỗi lịch sử không nên làm fail toàn bộ request
		history = []entity.Ownership{}
	}

	// 3. Lấy thông tin chủ sở hữu hiện tại từ User Module
	email, name, phone, err := s.userProvider.GetUserEmailByID(ctx, own.OwnerID)
	if err != nil {
		email, name, phone = "", "Unknown", ""
	}

	// 4. Lấy thông tin sản phẩm từ Product Module
	pName, sku, serialNumber, err := s.productPort.GetProductItemDetail(ctx, own.ProductItemID)
	if err != nil {
		pName, sku, serialNumber = "Unknown Product", "N/A", "N/A"
	}

	// 5. Map history — Mỗi record trong history cần fetch thêm user info
	historyItems := make([]dto.OwnershipHistoryItem, 0, len(history))
	for _, h := range history {
		hEmail, hName, hPhone, err := s.userProvider.GetUserEmailByID(ctx, h.OwnerID)
		if err != nil {
			hEmail, hName, hPhone = "", "Unknown", ""
		}
		historyItems = append(historyItems, dto.OwnershipHistoryItem{
			OwnershipID:      h.ID,
			OwnerName:        hName,
			OwnerEmail:       hEmail,
			OwnerPhone:       hPhone,
			Status:           string(h.Status),
			RegistrationDate: h.OwnedAt,
			EndedAt:          h.EndedAt,
		})
	}

	// 6. Trả về response đầy đủ
	return &dto.OwnershipDetailRes{
		OwnershipID:      own.ID,
		ProductID:        own.ProductItemID,
		Status:           string(own.Status),
		RegistrationDate: own.OwnedAt,
		OwnerID:          own.OwnerID,
		OwnerName:        name,
		OwnerEmail:       email,
		OwnerPhone:       phone,
		ProductName:      pName,
		ProductSKU:       sku,
		SerialNumber:     serialNumber,
		OwnershipHistory: historyItems,
	}, nil
}

// ---------------------------------------------------------------------------
// CRUD Extensions
// ---------------------------------------------------------------------------

// TransferOwnership (Update)
func (s *OwnershipService) TransferOwnership(ctx context.Context, id uuid.UUID, req dto.TransferOwnershipReq, currentUserID uuid.UUID, role string) error {
	oldOwnership, err := s.repo.GetOwnershipByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &apperror.AppError{
				Status:  404,
				Code:    apperror.CodeNotFound,
				Message: "Ownership không tồn tại",
			}
		}
		return err
	}

	// Authorization
	if role != "ADMIN" && oldOwnership.OwnerID != currentUserID {
		return &apperror.AppError{
			Status:  403,
			Code:    apperror.CodeForbidden,
			Message: "Bạn không có quyền thực hiện thao tác này.",
		}
	}

	if oldOwnership.Status != entity.OwnershipStatusActive {
		return &apperror.AppError{
			Status:  400,
			Code:    apperror.CodeBadRequest,
			Message: "Chỉ quyền sở hữu đang hoạt động mới có thể chuyển nhượng",
		}
	}

	// Ensure new user exists (auto-create if they don't, using their email, name, and phone)
	newOwnerID, err := s.userProvider.EnsureUserExists(ctx, req.NewOwnerEmail, req.NewOwnerName, req.NewOwnerPhone)
	if err != nil {
		return &apperror.AppError{
			Status:  400,
			Code:    apperror.CodeValidation,
			Message: "Không thể xác thực thông tin người sở hữu mới: " + err.Error(),
		}
	}

	// Cannot transfer to oneself
	if oldOwnership.OwnerID == newOwnerID {
		return &apperror.AppError{
			Status:  400,
			Code:    apperror.CodeValidation,
			Message: "Không thể chuyển nhượng cho chính mình",
		}
	}

	newOwnership := entity.NewOwnership(oldOwnership.ProductItemID, newOwnerID)
	newOwnership.OwnershipType = "TRANSFERRED"

	return s.repo.TransferOwnershipTx(ctx, oldOwnership.ID, newOwnership)
}

// DeleteOwnership (Soft Delete)
func (s *OwnershipService) DeleteOwnership(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, role string) error {
	oldOwnership, err := s.repo.GetOwnershipByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &apperror.AppError{
				Status:  404,
				Code:    apperror.CodeNotFound,
				Message: "Ownership không tồn tại",
			}
		}
		return err
	}

	if role != "ADMIN" && oldOwnership.OwnerID != currentUserID {
		return &apperror.AppError{
			Status:  403,
			Code:    apperror.CodeForbidden,
			Message: "Bạn không có quyền thực hiện thao tác này.",
		}
	}

	err = s.repo.UpdateOwnershipStatusAndEndedAt(ctx, nil, oldOwnership.ID, string(entity.OwnershipStatusRevoked))
	if err != nil {
		return err
	}

	return nil
}

// SearchOwnerships (List & Filter FR-040 & FR-042)
func (s *OwnershipService) SearchOwnerships(ctx context.Context, req dto.SearchOwnershipsReq, currentUserID uuid.UUID, role string) (*dto.PaginatedOwnershipsRes, error) {
	filter := repository.SearchFilter{
		Page:            req.Page,
		Limit:           req.Limit,
		OwnershipStatus: req.OwnershipStatus,
		Role:            role,
		CurrentUserID:   currentUserID,
	}

	// Cross-module search: Users
	if req.OwnerName != "" || req.OwnerEmail != "" || req.OwnerPhone != "" {
		ownerIDs, err := s.userProvider.SearchUserIDs(ctx, req.OwnerName, req.OwnerEmail, req.OwnerPhone)
		if err != nil {
			return nil, err
		}
		if len(ownerIDs) == 0 {
			return &dto.PaginatedOwnershipsRes{Data: []dto.OwnershipSummaryRes{}, TotalItems: 0, TotalPages: 0, Page: req.Page, Limit: req.Limit}, nil
		}
		filter.OwnerIDs = ownerIDs
	}

	// Cross-module search: Products
	if req.ProductName != "" || req.ProductCode != "" {
		itemIDs, err := s.productPort.SearchProductItemIDs(ctx, req.ProductName, req.ProductCode)
		if err != nil {
			return nil, err
		}
		if len(itemIDs) == 0 {
			return &dto.PaginatedOwnershipsRes{Data: []dto.OwnershipSummaryRes{}, TotalItems: 0, TotalPages: 0, Page: req.Page, Limit: req.Limit}, nil
		}
		filter.ProductItemIDs = itemIDs
	}

	ownerships, total, err := s.repo.SearchOwnerships(ctx, filter)
	if err != nil {
		return nil, err
	}

	results := make([]dto.OwnershipSummaryRes, len(ownerships))
	for i, o := range ownerships {
		results[i] = dto.OwnershipSummaryRes{
			OwnershipID:      o.ID,
			ProductID:        o.ProductItemID,
			Status:           string(o.Status),
			RegistrationDate: o.OwnedAt,
		}

		name, email, phone, _ := s.userProvider.GetUserEmailByID(ctx, o.OwnerID)
		results[i].OwnerName = name
		results[i].OwnerEmail = email
		results[i].OwnerPhone = phone

		prodName, prodSKU, prodSerial, _ := s.productPort.GetProductItemDetail(ctx, o.ProductItemID)
		results[i].ProductName = prodName
		results[i].ProductSKU = prodSKU
		results[i].SerialNumber = prodSerial
	}

	limit := req.Limit
	if limit < 1 {
		limit = 10
	}
	page := req.Page
	if page < 1 {
		page = 1
	}
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &dto.PaginatedOwnershipsRes{
		Data:       results,
		TotalItems: total,
		TotalPages: totalPages,
		Page:       page,
		Limit:      limit,
	}, nil
}

