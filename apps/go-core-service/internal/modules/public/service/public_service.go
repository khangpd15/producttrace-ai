package service

import (
	"context"
	"time"

	itemRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/public/dto"
)

type PublicService interface {
	VerifyQR(ctx context.Context, itemCode string, token string) (*dto.VerifyQRResponse, error)
}

type publicService struct {
	itemRepo itemRepo.ProductItemRepository
}

func NewPublicService(itemRepo itemRepo.ProductItemRepository) PublicService {
	return &publicService{
		itemRepo: itemRepo,
	}
}

func (s *publicService) VerifyQR(ctx context.Context, itemCode string, token string) (*dto.VerifyQRResponse, error) {
	row, err := s.itemRepo.FindByCodeAndToken(ctx, itemCode, token)
	if err != nil {
		return nil, err
	}

	return &dto.VerifyQRResponse{
		ItemCode:     row.ItemCode,
		SerialNumber: row.SerialNumber,
		Status:       row.ItemStatus,
		ScannedAt:    time.Now().UTC(),
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
			ProductName: row.ProductName,
			VariantName: row.VariantName,
			VariantSKU:  row.VariantSKU,
		},
	}, nil
}
