package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/entity"
	"gorm.io/gorm"
)

type WarrantyClaimRepository interface {
	Create(ctx context.Context, claim *entity.WarrantyClaim) error
	FindByProductItemIDAndStatusList(ctx context.Context, productItemID uuid.UUID, statuses []entity.WarrantyClaimStatus) (*entity.WarrantyClaim, error)
	FindByID(ctx context.Context, claimID uuid.UUID) (*entity.WarrantyClaim, error)
}

type gormWarrantyClaimRepository struct {
	db *gorm.DB
}

func NewWarrantyClaimRepository(db *gorm.DB) WarrantyClaimRepository {
	return &gormWarrantyClaimRepository{db: db}
}

func (r *gormWarrantyClaimRepository) Create(ctx context.Context, claim *entity.WarrantyClaim) error {
	return r.db.WithContext(ctx).Create(claim).Error
}

func (r *gormWarrantyClaimRepository) FindByProductItemIDAndStatusList(ctx context.Context, productItemID uuid.UUID, statuses []entity.WarrantyClaimStatus) (*entity.WarrantyClaim, error) {
	var claim entity.WarrantyClaim
	err := r.db.WithContext(ctx).Where("product_item_id = ? AND status IN ?", productItemID, statuses).First(&claim).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &claim, nil
}

func (r *gormWarrantyClaimRepository) FindByID(ctx context.Context, claimID uuid.UUID) (*entity.WarrantyClaim, error) {
	var claim entity.WarrantyClaim
	err := r.db.WithContext(ctx).Where("id = ?", claimID).First(&claim).Error
	if err != nil {
		return nil, err
	}
	return &claim, nil
}
