package repository

import (
    "context"
    "github.com/google/uuid"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/entity"
    "gorm.io/gorm"
)

type ProductRepository interface {
    Create(ctx context.Context, product *entity.Product) error
    Update(ctx context.Context, product *entity.Product) error
    FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)
    FindAll(ctx context.Context, filter ProductFilter) ([]entity.Product, int64, error)
    SoftDelete(ctx context.Context, id uuid.UUID) error
}

type ProductFilter struct {
    Search     *string
    CategoryID *uuid.UUID
    Status     *string
    Page       int
    Limit      int
    SortBy     *string
    SortOrder  *string
}

type productRepository struct {
    db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
    return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *entity.Product) error {
    return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepository) Update(ctx context.Context, product *entity.Product) error {
    return r.db.WithContext(ctx).Save(product).Error
}

func (r *productRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
    var product entity.Product
    err := r.db.WithContext(ctx).
        Preload("Category").
        Preload("Variants").
        Where("id = ? AND is_deleted = false", id).
        First(&product).Error
    if err != nil {
        return nil, err
    }
    return &product, nil
}

func (r *productRepository) FindAll(ctx context.Context, filter ProductFilter) ([]entity.Product, int64, error) {
    var products []entity.Product
    var total int64

    query := r.db.WithContext(ctx).
        Model(&entity.Product{}).
        Where("is_deleted = false")

    if filter.Search != nil {
        query = query.Where("name ILIKE ?", "%"+*filter.Search+"%")
    }
    if filter.CategoryID != nil {
        query = query.Where("category_id = ?", filter.CategoryID)
    }
    if filter.Status != nil {
        query = query.Where("status = ?", filter.Status)
    }

    query.Count(&total)

    if filter.SortBy != nil {
        order := "asc"
        if filter.SortOrder != nil {
            order = *filter.SortOrder
        }
        query = query.Order(*filter.SortBy + " " + order)
    }

    offset := (filter.Page - 1) * filter.Limit
    err := query.
        Preload("Category").
        Preload("Variants").
        Offset(offset).
        Limit(filter.Limit).
        Find(&products).Error

    return products, total, err
}

func (r *productRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
    return r.db.WithContext(ctx).
        Model(&entity.Product{}).
        Where("id = ?", id).
        Update("is_deleted", true).Error
}