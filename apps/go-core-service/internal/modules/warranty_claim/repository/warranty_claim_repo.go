package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/entity"
	"gorm.io/gorm"
)

type WarrantyClaimRepository interface {
	Create(ctx context.Context, claim *entity.WarrantyClaim) error
	Update(ctx context.Context, claim *entity.WarrantyClaim) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.WarrantyClaim, error)
	FindAll(ctx context.Context) ([]entity.WarrantyClaim, error)
	FindByWarrantyID(ctx context.Context, warrantyID uuid.UUID) ([]entity.WarrantyClaim, error)
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type warrantyClaimRepo struct {
	db *gorm.DB
}

func NewWarrantyClaimRepository(db *gorm.DB) WarrantyClaimRepository {
	return &warrantyClaimRepo{db: db}
}

func (r *warrantyClaimRepo) Create(ctx context.Context, claim *entity.WarrantyClaim) error {
	return r.db.WithContext(ctx).Create(claim).Error
}

func (r *warrantyClaimRepo) Update(ctx context.Context, claim *entity.WarrantyClaim) error {
	return r.db.WithContext(ctx).Save(claim).Error
}

func (r *warrantyClaimRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.WarrantyClaim, error) {
	var claim entity.WarrantyClaim
	err := r.db.WithContext(ctx).First(&claim, "id = ?", id).Error
	return &claim, err
}

func (r *warrantyClaimRepo) FindAll(ctx context.Context) ([]entity.WarrantyClaim, error) {
	var claims []entity.WarrantyClaim
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&claims).Error
	return claims, err
}

func (r *warrantyClaimRepo) FindByWarrantyID(ctx context.Context, warrantyID uuid.UUID) ([]entity.WarrantyClaim, error) {
	var claims []entity.WarrantyClaim
	err := r.db.WithContext(ctx).Where("warranty_id = ?", warrantyID).Order("created_at desc").Find(&claims).Error
	return claims, err
}

func (r *warrantyClaimRepo) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}
