package repositories

import (
	"context"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/entities"
	"gorm.io/gorm"
)

type ProductVariantRepository interface {
	Create(ctx context.Context, variant *entities.ProductVariant) error
	ExistsBySKU(ctx context.Context, sku string) (bool, error)
	ExistsByBarcode(ctx context.Context, barcode string) (bool, error)
}

type productVariantRepository struct {
	db *gorm.DB
}

func NewProductVariantRepository(db *gorm.DB) ProductVariantRepository {
	return &productVariantRepository{db: db}
}

func (r *productVariantRepository) Create(ctx context.Context, variant *entities.ProductVariant) error {
	return GetDB(ctx, r.db).Create(variant).Error
}

func (r *productVariantRepository) ExistsBySKU(ctx context.Context, sku string) (bool, error) {
	var count int64
	err := GetDB(ctx, r.db).Model(&entities.ProductVariant{}).
		Where("sku = ? AND is_deleted = ?", sku, false).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *productVariantRepository) ExistsByBarcode(ctx context.Context, barcode string) (bool, error) {
	var count int64
	err := GetDB(ctx, r.db).Model(&entities.ProductVariant{}).
		Where("barcode = ? AND is_deleted = ?", barcode, false).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
