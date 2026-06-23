package repositories

import (
    "context"
    "github.com/google/uuid"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/entities"
    "gorm.io/gorm"
)

type ProductCategoryRepository interface {
    Create(ctx context.Context, category *entities.ProductCategory) error
    ExistsByName(ctx context.Context, name string) (bool, error)
    ExistsByCode(ctx context.Context, code string) (bool, error)
    ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}

type productCategoryRepository struct {
    db *gorm.DB
}

func NewProductCategoryRepository(db *gorm.DB) ProductCategoryRepository {
    return &productCategoryRepository{db: db}
}

func (r *productCategoryRepository) Create(ctx context.Context, category *entities.ProductCategory) error {
    return r.db.WithContext(ctx).Create(category).Error
}

func (r *productCategoryRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&entities.ProductCategory{}).
        Where("name = ?", name).
        Count(&count).Error
    return count > 0, err
}

func (r *productCategoryRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&entities.ProductCategory{}).
        Where("code = ?", code).
        Count(&count).Error
    return count > 0, err
}

func (r *productCategoryRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&entities.ProductCategory{}).
        Where("id = ?", id).
        Count(&count).Error
    return count > 0, err
}