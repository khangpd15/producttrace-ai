package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/entity"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/repository"
	"gorm.io/gorm"
)

type IOwnershipService interface {
	CustomerRequestOTP(ctx context.Context, req dto.CustomerRequestOTPReq, userID uuid.UUID) error
	CustomerVerifyAndRegister(ctx context.Context, req dto.CustomerVerifyAndRegisterReq, userID uuid.UUID) (*entity.Ownership, error)

	AdminRequestOTP(ctx context.Context, req dto.AdminRequestOTPReq, adminID uuid.UUID) error
	AdminVerifyAndRegister(ctx context.Context, req dto.AdminVerifyAndRegisterReq, adminID uuid.UUID) (*entity.Ownership, error)

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
}

func NewOwnershipService(
	repo repository.IOwnershipRepository,
	productPort IProductService,
	emailClient IEmailOTPClient,
	userProvider IUserInfoProvider,
) IOwnershipService {
	return &OwnershipService{
		repo:         repo,
		productPort:  productPort,
		emailClient:  emailClient,
		userProvider: userProvider,
	}
}

// ---------------------------------------------------------------------------
// CUSTOMER FLOW
// ---------------------------------------------------------------------------

func (s *OwnershipService) CustomerRequestOTP(ctx context.Context, req dto.CustomerRequestOTPReq, userID uuid.UUID) error {
	email, _, _, err := s.userProvider.GetUserEmailByID(ctx, userID)
	if err != nil {
		return err
	}

	productID, err := s.productPort.FindProductByQR(ctx, req.QRCode)
	if err != nil {
		return err
	}
	if err := s.productPort.ValidateProductOwnershipStatus(ctx, productID); err != nil {
		return err
	}

	return s.emailClient.RequestOTP(ctx, email, productID.String())
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

	newOwnership := entity.NewOwnership(req.ProductID, userID)
	saved, err := s.repo.CreateOwnership(ctx, nil, newOwnership)
	if err != nil {
		return nil, err
	}

	if err := s.productPort.UpdateOwnershipStatus(ctx, req.ProductID, string(entity.OwnershipStatusActive)); err != nil {
		return nil, err
	}

	return saved, nil
}

// ---------------------------------------------------------------------------
// ADMIN FLOW
// ---------------------------------------------------------------------------

func (s *OwnershipService) AdminRequestOTP(ctx context.Context, req dto.AdminRequestOTPReq, adminID uuid.UUID) error {
	productID, err := s.productPort.FindProductByQR(ctx, req.QRCode)
	if err != nil {
		return err
	}
	if err := s.productPort.ValidateProductOwnershipStatus(ctx, productID); err != nil {
		return err
	}

	return s.emailClient.RequestOTP(ctx, req.OwnerEmail, productID.String())
}

func (s *OwnershipService) AdminVerifyAndRegister(ctx context.Context, req dto.AdminVerifyAndRegisterReq, adminID uuid.UUID) (*entity.Ownership, error) {
	valid, err := s.emailClient.ValidateOTP(ctx, req.OwnerEmail, req.OTP)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, errors.New("invalid or expired OTP")
	}

	ownerID, err := s.userProvider.EnsureUserExists(ctx, req.OwnerEmail, req.OwnerName, req.OwnerPhone)
	if err != nil {
		return nil, err
	}

	newOwnership := entity.NewOwnership(req.ProductID, ownerID)
	saved, err := s.repo.CreateOwnership(ctx, nil, newOwnership)
	if err != nil {
		return nil, err
	}

	if err := s.productPort.UpdateOwnershipStatus(ctx, req.ProductID, string(entity.OwnershipStatusActive)); err != nil {
		return nil, err
	}

	return saved, nil
}

// ---------------------------------------------------------------------------
// DETAIL API (UC-P1-OWNER-02)
// ---------------------------------------------------------------------------

func (s *OwnershipService) GetOwnershipDetail(ctx context.Context, productItemID uuid.UUID) (*dto.OwnershipDetailRes, error) {
	// 1. Lấy record sở hữu hiện tại (mới nhất) — VAL-002: phân biệt "ownership not found"
	own, err := s.repo.GetOwnershipByProductItemID(ctx, productItemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ownership information not found")
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
	pName, sku, err := s.productPort.GetProductItemDetail(ctx, own.ProductItemID)
	if err != nil {
		pName, sku = "Unknown Product", "N/A"
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
			return errors.New("ownership information not found")
		}
		return err
	}

	// Authorization
	if role != "ADMIN" && oldOwnership.OwnerID != currentUserID {
		return errors.New("access denied")
	}

	if oldOwnership.Status != entity.OwnershipStatusActive {
		return errors.New("only active ownerships can be transferred")
	}

	// Ensure new user exists
	newOwnerID, err := s.userProvider.EnsureUserExists(ctx, req.NewOwnerEmail, req.NewOwnerName, req.NewOwnerPhone)
	if err != nil {
		return err
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
			return errors.New("ownership information not found")
		}
		return err
	}

	if role != "ADMIN" && oldOwnership.OwnerID != currentUserID {
		return errors.New("access denied")
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

		prodName, prodSKU, _ := s.productPort.GetProductItemDetail(ctx, o.ProductItemID)
		results[i].ProductName = prodName
		results[i].ProductSKU = prodSKU
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

