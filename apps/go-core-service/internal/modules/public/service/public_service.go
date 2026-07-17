package service

import (
	"context"
	"time"

	itemRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/public/dto"
	traceRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/trace/repositories"
	ownershipRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/repository"
	warrantyRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/repository"
	locationRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/repository"
	userRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/repository"
)

type PublicService interface {
	VerifyQR(ctx context.Context, itemCode string, token string) (*dto.VerifyQRResponse, error)
}

type publicService struct {
	itemRepo      itemRepo.ProductItemRepository
	traceRepo     traceRepo.TraceRepository
	ownershipRepo ownershipRepo.IOwnershipRepository
	warrantyRepo  warrantyRepo.WarrantyRepository
	locationRepo  locationRepo.LocationRepository
	userRepo      userRepo.UserRepositoryInterface
}

func NewPublicService(
	itemRepo itemRepo.ProductItemRepository,
	traceRepo traceRepo.TraceRepository,
	ownershipRepo ownershipRepo.IOwnershipRepository,
	warrantyRepo warrantyRepo.WarrantyRepository,
	locationRepo locationRepo.LocationRepository,
	userRepo userRepo.UserRepositoryInterface,
) PublicService {
	return &publicService{
		itemRepo:      itemRepo,
		traceRepo:     traceRepo,
		ownershipRepo: ownershipRepo,
		warrantyRepo:  warrantyRepo,
		locationRepo:  locationRepo,
		userRepo:      userRepo,
	}
}

func (s *publicService) VerifyQR(ctx context.Context, itemCode string, token string) (*dto.VerifyQRResponse, error) {
	row, err := s.itemRepo.FindByCodeAndToken(ctx, itemCode, token)
	if err != nil {
		return nil, err
	}

	// 1. Fetch Events
	var traceHistory []dto.VerifyQREvent
	timelineEvents, err := s.traceRepo.FindEvents(ctx, row.ID, row.BatchID, nil, nil, nil)
	if err == nil {
		for _, e := range timelineEvents {
			traceHistory = append(traceHistory, dto.VerifyQREvent{
				EventType:   e.EventType,
				Title:       e.Title,
				Description: e.Description,
				Location:    e.Location,
				ActorName:   e.ActorName,
				OccurredAt:  e.Timestamp,
			})
		}
	}

	// 2. Fetch Ownership
	var verifyOwnership *dto.VerifyQROwnership
	ownership, err := s.ownershipRepo.GetOwnershipByProductItemID(ctx, row.ID)
	if err == nil && ownership != nil {
		ownerName := "Unknown User"
		ownerUser, err := s.userRepo.GetUserByID(ctx, ownership.OwnerID.String())
		if err == nil && ownerUser != nil {
			ownerName = ownerUser.FullName
		}
		verifyOwnership = &dto.VerifyQROwnership{
			OwnerName:     ownerName,
			RegisteredAt:  ownership.OwnedAt,
			OwnershipType: ownership.OwnershipType,
			Status:        string(ownership.Status),
		}
	}

	// 3. Fetch Warranty
	var verifyWarranty *dto.VerifyQRWarranty
	warranties, err := s.warrantyRepo.FindBySerialNumber(ctx, row.SerialNumber)
	if err == nil && len(warranties) > 0 {
		verifyWarranty = &dto.VerifyQRWarranty{
			ClaimNumber: warranties[0].WarrantyCode,
			Status:      string(warranties[0].Status),
			CreatedAt:   warranties[0].CreatedAt,
		}
	}

	// 4. Fetch Location
	var verifyLocation *dto.VerifyQRLocation
	if row.CurrentLocationID != nil {
		loc, err := s.locationRepo.GetByID(ctx, row.CurrentLocationID.String())
		if err == nil && loc != nil {
			verifyLocation = &dto.VerifyQRLocation{
				Name:    loc.Name,
				Type:    string(loc.Type),
				Address: loc.Address,
				City:    loc.City,
			}
		}
	}

	desc := ""
	if row.ProductDescription != nil {
		desc = *row.ProductDescription
	}

	thumbnail := ""
	if row.ProductThumbnailURL != nil {
		thumbnail = *row.ProductThumbnailURL
	}

	barcode := ""
	if row.VariantBarcode != nil {
		barcode = *row.VariantBarcode
	}

	ownershipStatus := ""
	if verifyOwnership != nil {
		ownershipStatus = verifyOwnership.Status
	}

	warrantyStatus := ""
	if verifyWarranty != nil {
		warrantyStatus = verifyWarranty.Status
	}

	if traceHistory == nil {
		traceHistory = []dto.VerifyQREvent{}
	}

	return &dto.VerifyQRResponse{
		ProductItemID:   row.ID,
		ProductID:       row.ProductID,
		VariantID:       row.VariantID,
		BatchID:         row.BatchID,
		ItemCode:        row.ItemCode,
		SerialNumber:    row.SerialNumber,
		Status:          row.ItemStatus,
		ScannedAt:       time.Now().UTC(),
		OwnershipStatus: ownershipStatus,
		WarrantyStatus:  warrantyStatus,
		Batch: dto.VerifyQRBatchInfo{
			BatchCode:        row.BatchCode,
			ManufactureDate:  row.ManufactureDate,
			ExpiryDate:       row.ExpiryDate,
			ManufacturerName: row.ManufacturerName,
			SupplierName:     row.SupplierName,
			OriginCountry:    row.OriginCountry,
			ProductionPlace:  row.ProductionPlace,
			Status:           row.BatchStatus,
		},
		Product: dto.VerifyQRProductInfo{
			ProductName:  row.ProductName,
			Description:  desc,
			ThumbnailURL: thumbnail,
			CategoryName: row.CategoryName,
			VariantName:  row.VariantName,
			VariantSKU:   row.VariantSKU,
			Barcode:      barcode,
		},
		Ownership:    verifyOwnership,
		Warranty:     verifyWarranty,
		Location:     verifyLocation,
		TraceHistory: traceHistory,
	}, nil
}
