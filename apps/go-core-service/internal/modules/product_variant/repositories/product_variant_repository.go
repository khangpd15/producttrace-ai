package repositories

import (
    "context"

    "github.com/google/uuid"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/entities"
    "gorm.io/gorm"
)

type ProductVariantRepository interface {
    Create(ctx context.Context, variant *entities.ProductVariant) error
    ExistsBySKU(ctx context.Context, sku string) (bool, error)
    ExistsByBarcode(ctx context.Context, barcode string) (bool, error)
    ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}

type txKey struct{}

func InjectTx(ctx context.Context, tx *gorm.DB) context.Context {
    return context.WithValue(ctx, txKey{}, tx)
}

func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
    if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
        return tx.WithContext(ctx)
    }
    return defaultDB.WithContext(ctx)
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
    err := GetDB(ctx, r.db).
        Model(&entities.ProductVariant{}).
        Where("sku = ? AND is_deleted = false", sku).
        Count(&count).Error
    return count > 0, err
}

func (r *productVariantRepository) ExistsByBarcode(ctx context.Context, barcode string) (bool, error) {
    var count int64
    err := GetDB(ctx, r.db).
        Model(&entities.ProductVariant{}).
        Where("barcode = ? AND is_deleted = false", barcode).
        Count(&count).Error
    return count > 0, err
}

// ExistsByID kiểm tra variant có tồn tại và chưa bị xóa mềm.
// Dùng COUNT(1) qua GORM để tránh SELECT *.
func (r *productVariantRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
    var count int64
    err := GetDB(ctx, r.db).
        Model(&entities.ProductVariant{}).
        Where("id = ? AND is_deleted = false", id).
        Count(&count).Error
    return count > 0, err
}