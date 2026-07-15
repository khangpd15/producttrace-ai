package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/entity"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/repository"
	"gorm.io/gorm"
)

var (
	ErrWarrantyExists = errors.New("warranty already exists for this product")
	ErrWarrantyNotFound = errors.New("warranty not found")
	ErrNotOwned = errors.New("product is not registered for ownership or ownership is inactive")
	ErrOver30Days = errors.New("warranty registration request exceeds 30 days from ownership date")
)

type WarrantyService interface {
	ActivateWarranty(ctx context.Context, req dto.CreateWarrantyRequest) (*dto.WarrantyResponse, error)
	RequestWarranty(ctx context.Context, req dto.CustomerRequestWarrantyRequest) (*dto.WarrantyResponse, error)
	ApproveWarranty(ctx context.Context, id uuid.UUID, req dto.ApproveWarrantyRequest) (*dto.WarrantyResponse, error)
	RejectWarranty(ctx context.Context, id uuid.UUID, req dto.RejectWarrantyRequest) (*dto.WarrantyResponse, error)
	ListWarranties(ctx context.Context) ([]dto.WarrantyResponse, error)
	GetWarrantyByCode(ctx context.Context, code string) (*dto.WarrantyResponse, error)
}

type warrantyService struct {
	repo repository.WarrantyRepository
}

func NewWarrantyService(repo repository.WarrantyRepository) WarrantyService {
	return &warrantyService{repo: repo}
}

func (s *warrantyService) ActivateWarranty(ctx context.Context, req dto.CreateWarrantyRequest) (*dto.WarrantyResponse, error) {
	// ... keeping this for backward compatibility if needed, or Admin directly creates ...
	// 1. Check if warranty code already exists
	_, err := s.repo.FindByWarrantyCode(ctx, req.WarrantyCode)
	if err == nil {
		return nil, ErrWarrantyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 2. Parse dates
	var startDate, endDate time.Time
	if req.StartDate != "" {
		startDate, _ = time.Parse("2006-01-02", req.StartDate)
	} else {
		startDate = time.Now()
	}
	
	if req.EndDate != "" {
		endDate, _ = time.Parse("2006-01-02", req.EndDate)
	} else {
		endDate = startDate.AddDate(0, req.DurationMonths, 0)
	}

	status := entity.WarrantyStatusActive
	if req.Status != "" {
		status = entity.WarrantyStatus(req.Status)
	}

	// 3. Create entity
	warranty := &entity.Warranty{
		ID:                uuid.New(),
		ItemCode:          req.ItemCode,
		ItemName:          req.ItemName,
		SerialNumber:      req.SerialNumber,
		OwnerName:         req.OwnerName,
		OwnerEmail:        req.OwnerEmail,
		WarrantyCode:      req.WarrantyCode,
		PolicyName:        req.PolicyName,
		PolicyDescription: req.PolicyDescription,
		DurationMonths:    req.DurationMonths,
		Status:            status,
		StartDate:         startDate,
		EndDate:           endDate,
		InvoiceNumber:     req.InvoiceNumber,
		Note:              req.Note,
	}

	if err := s.repo.Create(ctx, warranty); err != nil {
		return nil, err
	}

	resp := dto.FromWarrantyEntity(warranty)
	return &resp, nil
}

func (s *warrantyService) RequestWarranty(ctx context.Context, req dto.CustomerRequestWarrantyRequest) (*dto.WarrantyResponse, error) {
	// 1. Check Ownership
	_, err := s.repo.GetOwnershipDate(ctx, req.ItemCode, req.SerialNumber)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotOwned
		}
		return nil, err
	}

	// 2. Check if a warranty already exists for this serial number
	existing, err := s.repo.FindBySerialNumber(ctx, req.SerialNumber)
	if err == nil && len(existing) > 0 {
		return nil, ErrWarrantyExists
	}

	// 3. Create Pending Warranty
	warranty := &entity.Warranty{
		ID:                uuid.New(),
		ItemCode:          req.ItemCode,
		SerialNumber:      req.SerialNumber,
		OwnerName:         req.OwnerName,
		OwnerEmail:        req.OwnerEmail,
		WarrantyCode:      "WAR-" + uuid.New().String()[:8], // Generate a random code or wait for admin
		Status:            entity.WarrantyStatusPending,
		Note:              req.Note,
	}

	if err := s.repo.Create(ctx, warranty); err != nil {
		return nil, err
	}

	resp := dto.FromWarrantyEntity(warranty)
	return &resp, nil
}

func (s *warrantyService) ApproveWarranty(ctx context.Context, id uuid.UUID, req dto.ApproveWarrantyRequest) (*dto.WarrantyResponse, error) {
	warranty, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWarrantyNotFound
		}
		return nil, err
	}

	if warranty.Status != entity.WarrantyStatusPending {
		return nil, errors.New("only pending warranties can be approved")
	}

	ownedAt, err := s.repo.GetOwnershipDate(ctx, warranty.ItemCode, warranty.SerialNumber)
	if err != nil {
		return nil, errors.New("could not verify ownership")
	}

	if time.Since(*ownedAt) > 30*24*time.Hour {
		return nil, ErrOver30Days
	}

	startDate := time.Now()
	endDate := startDate.AddDate(0, req.DurationMonths, 0)

	warranty.Status = entity.WarrantyStatusActive
	warranty.PolicyName = req.PolicyName
	warranty.DurationMonths = req.DurationMonths
	warranty.StartDate = startDate
	warranty.EndDate = endDate

	if err := s.repo.Update(ctx, warranty); err != nil {
		return nil, err
	}

	resp := dto.FromWarrantyEntity(warranty)
	return &resp, nil
}

func (s *warrantyService) RejectWarranty(ctx context.Context, id uuid.UUID, req dto.RejectWarrantyRequest) (*dto.WarrantyResponse, error) {
	warranty, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWarrantyNotFound
		}
		return nil, err
	}

	if warranty.Status != entity.WarrantyStatusPending {
		return nil, errors.New("only pending warranties can be rejected")
	}

	warranty.Status = entity.WarrantyStatusRejected
	warranty.Note = req.Reason

	if err := s.repo.Update(ctx, warranty); err != nil {
		return nil, err
	}

	resp := dto.FromWarrantyEntity(warranty)
	return &resp, nil
}

func (s *warrantyService) ListWarranties(ctx context.Context) ([]dto.WarrantyResponse, error) {
	warranties, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	return dto.FromWarrantyEntities(warranties), nil
}

func (s *warrantyService) GetWarrantyByCode(ctx context.Context, code string) (*dto.WarrantyResponse, error) {
	warranty, err := s.repo.FindByWarrantyCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWarrantyNotFound
		}
		return nil, err
	}
	resp := dto.FromWarrantyEntity(warranty)
	return &resp, nil
}
