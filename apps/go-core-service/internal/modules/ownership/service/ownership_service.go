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
