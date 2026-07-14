package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/dbctx"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entities.Product) error
	Update(ctx context.Context, product *entities.Product) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Product, error)
	FindAll(ctx context.Context, filter ProductFilter) ([]entities.Product, int64, error)
	FindAllWithStats(ctx context.Context, filter ProductFilter) ([]response.ProductListItemDTO, int64, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	HasProductsByOwner(ctx context.Context, ownerID uuid.UUID) (bool, error)
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

// InjectTx / GetDB giờ dùng chung từ package pkg/dbctx (xem dbctx.go)
// để đảm bảo product, product_variant, product_attribute_value cùng
// nhận diện đúng 1 transaction khi cascade create/delete.

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *entities.Product) error {
	return dbctx.GetDB(ctx, r.db).Create(product).Error
}

func (r *productRepository) Update(ctx context.Context, product *entities.Product) error {
	return dbctx.GetDB(ctx, r.db).Save(product).Error
}

func (r *productRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	var product entities.Product
	err := dbctx.GetDB(ctx, r.db).
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

	query := dbctx.GetDB(ctx, r.db).
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

func (r *productRepository) FindAllWithStats(ctx context.Context, filter ProductFilter) ([]response.ProductListItemDTO, int64, error) {
	var items []response.ProductListItemDTO
	var total int64

	query := dbctx.GetDB(ctx, r.db).
		Table("products p").
		Where("p.is_deleted = false")

	if filter.Search != nil {
		query = query.Where("p.name ILIKE ?", "%"+*filter.Search+"%")
	}
	if filter.CategoryID != nil {
		query = query.Where("p.category_id = ?", filter.CategoryID)
	}
	if filter.Status != nil {
		query = query.Where("p.status = ?", filter.Status)
	}

	query.Count(&total)

	offset := (filter.Page - 1) * filter.Limit

	err := query.
		Select(`
			p.id, p.name, 
			COALESCE(c.name, '') as category_name, 
			p.status, p.created_at, p.thumbnail_url,
			(SELECT COUNT(*) FROM product_variants pv WHERE pv.product_id = p.id AND pv.is_deleted = false) as variants_count,
			(SELECT COUNT(*) FROM batches b JOIN product_variants pv ON b.variant_id = pv.id WHERE pv.product_id = p.id AND b.is_deleted = false) as batches_count
		`).
		Joins("LEFT JOIN product_categories c ON p.category_id = c.id").
		Order("p.created_at DESC").
		Offset(offset).
		Limit(filter.Limit).
		Scan(&items).Error

	return items, total, err
}

func (r *productRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return dbctx.GetDB(ctx, r.db).
		Model(&entities.Product{}).
		Where("id = ?", id).
		Update("is_deleted", true).Error
}

func (r *productRepository) HasProductsByOwner(ctx context.Context, ownerID uuid.UUID) (bool, error) {
	var count int64
	err := GetDB(ctx, r.db).
		Model(&entities.Product{}).
		Where("created_by = ? AND is_deleted = false", ownerID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
