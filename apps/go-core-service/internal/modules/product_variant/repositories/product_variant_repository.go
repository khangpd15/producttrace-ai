package repositories

import (
    "context"
    "github.com/google/uuid"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/entities"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/dbctx"
    "gorm.io/gorm"
)

type ProductVariantRepository interface {
    Create(ctx context.Context, variant *entities.ProductVariant) error
    Update(ctx context.Context, variant *entities.ProductVariant) error
    FindByID(ctx context.Context, id uuid.UUID) (*entities.ProductVariant, error)
    FindByProductID(ctx context.Context, productID uuid.UUID) ([]entities.ProductVariant, error)
    SoftDelete(ctx context.Context, id uuid.UUID) error
    ExistsBySKU(ctx context.Context, sku string) (bool, error)
    ExistsByBarcode(ctx context.Context, barcode string) (bool, error)
    ExistsBySKUExcludeID(ctx context.Context, sku string, id uuid.UUID) (bool, error)
    ExistsByBarcodeExcludeID(ctx context.Context, barcode string, id uuid.UUID) (bool, error)
    ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
    SoftDeleteByProductID(ctx context.Context, productID uuid.UUID) error
}

// InjectTx / GetDB giờ dùng chung từ package pkg/dbctx (xem dbctx.go).

type productVariantRepository struct {
	db *gorm.DB
}

func NewProductVariantRepository(db *gorm.DB) ProductVariantRepository {
	return &productVariantRepository{db: db}
}

func (r *productVariantRepository) Create(ctx context.Context, variant *entities.ProductVariant) error {
	return dbctx.GetDB(ctx, r.db).Create(variant).Error
}

func (r *productVariantRepository) Update(ctx context.Context, variant *entities.ProductVariant) error {
    return dbctx.GetDB(ctx, r.db).Save(variant).Error
}

func (r *productVariantRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.ProductVariant, error) {
    var variant entities.ProductVariant
    err := dbctx.GetDB(ctx, r.db).
        Where("id = ? AND is_deleted = false", id).
        First(&variant).Error
    if err != nil {
        return nil, err
    }
    return &variant, nil
}

func (r *productVariantRepository) FindByProductID(ctx context.Context, productID uuid.UUID) ([]entities.ProductVariant, error) {
    var variants []entities.ProductVariant
    err := dbctx.GetDB(ctx, r.db).
        Where("product_id = ? AND is_deleted = false", productID).
        Find(&variants).Error
    return variants, err
}

func (r *productVariantRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
    return dbctx.GetDB(ctx, r.db).
        Model(&entities.ProductVariant{}).
        Where("id = ?", id).
        Update("is_deleted", true).Error
}

func (r *productVariantRepository) ExistsBySKU(ctx context.Context, sku string) (bool, error) {
	var count int64
	err := dbctx.GetDB(ctx, r.db).
		Model(&entities.ProductVariant{}).
		Where("sku = ? AND is_deleted = false", sku).
		Count(&count).Error
	return count > 0, err
}

func (r *productVariantRepository) ExistsByBarcode(ctx context.Context, barcode string) (bool, error) {
	var count int64
	err := dbctx.GetDB(ctx, r.db).
		Model(&entities.ProductVariant{}).
		Where("barcode = ? AND is_deleted = false", barcode).
		Count(&count).Error
	return count > 0, err
}

func (r *productVariantRepository) ExistsBySKUExcludeID(ctx context.Context, sku string, id uuid.UUID) (bool, error) {
    var count int64
    err := dbctx.GetDB(ctx, r.db).
        Model(&entities.ProductVariant{}).
        Where("sku = ? AND id != ? AND is_deleted = false", sku, id).
        Count(&count).Error
    return count > 0, err
}

func (r *productVariantRepository) ExistsByBarcodeExcludeID(ctx context.Context, barcode string, id uuid.UUID) (bool, error) {
    var count int64
    err := dbctx.GetDB(ctx, r.db).
        Model(&entities.ProductVariant{}).
        Where("barcode = ? AND id != ? AND is_deleted = false", barcode, id).
        Count(&count).Error
    return count > 0, err
}

func (r *productVariantRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := dbctx.GetDB(ctx, r.db).
		Model(&entities.ProductVariant{}).
		Where("id = ? AND is_deleted = false", id).
		Count(&count).Error
	return count > 0, err
}

func (r *productVariantRepository) SoftDeleteByProductID(ctx context.Context, productID uuid.UUID) error {
	return dbctx.GetDB(ctx, r.db).
		Model(&entities.ProductVariant{}).
		Where("product_id = ?", productID).
		Update("is_deleted", true).Error
}