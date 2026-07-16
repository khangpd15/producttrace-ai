package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	warrantyRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/repository"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/entity"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/repository"
	"gorm.io/gorm"
)

var (
	ErrNoActiveWarranty = errors.New("không tìm thấy hợp đồng bảo hành nào đang hoạt động cho sản phẩm này")
	ErrClaimNotFound    = errors.New("không tìm thấy yêu cầu bảo hành")
	ErrInvalidStatus    = errors.New("trạng thái yêu cầu bảo hành không hợp lệ")
)

type WarrantyClaimService interface {
	CustomerCreateClaim(ctx context.Context, req dto.CreateClaimRequest) (*dto.WarrantyClaimResponse, error)
	AdminUpdateClaimStatus(ctx context.Context, id uuid.UUID, req dto.UpdateClaimStatusRequest) (*dto.WarrantyClaimResponse, error)
	GetClaimByID(ctx context.Context, id uuid.UUID) (*dto.WarrantyClaimResponse, error)
	ListAllClaims(ctx context.Context) ([]dto.WarrantyClaimResponse, error)
	ListMyClaims(ctx context.Context, userID uuid.UUID) ([]dto.WarrantyClaimResponse, error)
}

type warrantyClaimService struct {
	repo         repository.WarrantyClaimRepository
	warrantyRepo warrantyRepo.WarrantyRepository
}

func NewWarrantyClaimService(repo repository.WarrantyClaimRepository, wRepo warrantyRepo.WarrantyRepository) WarrantyClaimService {
	return &warrantyClaimService{repo: repo, warrantyRepo: wRepo}
}

func (s *warrantyClaimService) CustomerCreateClaim(ctx context.Context, req dto.CreateClaimRequest) (*dto.WarrantyClaimResponse, error) {
	// 1. Find active warranty by serial number
	warranties, err := s.warrantyRepo.FindBySerialNumber(ctx, req.SerialNumber)
	if err != nil || len(warranties) == 0 {
		return nil, ErrNoActiveWarranty
	}

	var activeWarranty *uuid.UUID
	var productItem *uuid.UUID
	for _, w := range warranties {
		if string(w.Status) == "ACTIVE" {
			activeWarranty = &w.ID
			productItem = &w.ProductItemID
			break
		}
	}

	if activeWarranty == nil {
		return nil, ErrNoActiveWarranty
	}

	// 2. Create the claim
	claim := &entity.WarrantyClaim{
		ID:               uuid.New(),
		WarrantyID:       *activeWarranty,
		ProductItemID:    *productItem,
		CustomerName:     "Guest", // Or resolve from auth context later if needed
		CustomerPhone:    req.ContactPhone,
		CustomerEmail:    req.ContactEmail,
		IssueTitle:       req.IssueTitle,
		IssueDescription: req.IssueDescription,
		Status:           entity.ClaimStatusPending,
	}

	if err := s.repo.Create(ctx, claim); err != nil {
		return nil, err
	}

	resp := dto.FromEntity(claim)
	return &resp, nil
}

func (s *warrantyClaimService) AdminUpdateClaimStatus(ctx context.Context, id uuid.UUID, req dto.UpdateClaimStatusRequest) (*dto.WarrantyClaimResponse, error) {
	claim, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClaimNotFound
		}
		return nil, err
	}

	// Validate status transition (simplified)
	if req.Status == "" {
		return nil, ErrInvalidStatus
	}

	claim.Status = entity.ClaimStatus(req.Status)
	claim.ResolutionNote = req.ResolutionNote

	if err := s.repo.Update(ctx, claim); err != nil {
		return nil, err
	}

	resp := dto.FromEntity(claim)
	return &resp, nil
}

func (s *warrantyClaimService) GetClaimByID(ctx context.Context, id uuid.UUID) (*dto.WarrantyClaimResponse, error) {
	claim, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClaimNotFound
		}
		return nil, err
	}
	resp := dto.FromEntity(claim)
	return &resp, nil
}

func (s *warrantyClaimService) ListAllClaims(ctx context.Context) ([]dto.WarrantyClaimResponse, error) {
	claims, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return dto.FromEntities(claims), nil
}

func (s *warrantyClaimService) ListMyClaims(ctx context.Context, userID uuid.UUID) ([]dto.WarrantyClaimResponse, error) {
	// 1. Fetch user's warranties
	warranties, err := s.warrantyRepo.FindMyWarranties(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. Fetch claims for these warranties
	var allClaims []entity.WarrantyClaim
	for _, w := range warranties {
		claims, err := s.repo.FindByWarrantyID(ctx, w.ID)
		if err == nil {
			allClaims = append(allClaims, claims...)
		}
	}

	return dto.FromEntities(allClaims), nil
}
