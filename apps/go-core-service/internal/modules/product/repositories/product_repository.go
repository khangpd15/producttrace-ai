package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/entities"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entities.Product) error
	Update(ctx context.Context, product *entities.Product) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Product, error)
	FindAll(ctx context.Context, filter ProductFilter) ([]entities.Product, int64, error)
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

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *entities.Product) error {
	return GetDB(ctx, r.db).Create(product).Error
}

func (r *productRepository) Update(ctx context.Context, product *entities.Product) error {
	return GetDB(ctx, r.db).Save(product).Error
}

func (r *productRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	var product entities.Product
	err := GetDB(ctx, r.db).
		Preload("Category").
		Preload("Variants").
		Where("id = ? AND is_deleted = false", id).
		First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) FindAll(ctx context.Context, filter ProductFilter) ([]entities.Product, int64, error) {
	var products []entities.Product
	var total int64

	query := GetDB(ctx, r.db).
		Model(&entities.Product{}).
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
	return GetDB(ctx, r.db).
		Model(&entities.Product{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}
