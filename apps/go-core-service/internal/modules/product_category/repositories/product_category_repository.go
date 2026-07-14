package repositories

import (
	"context"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/entities"
	"gorm.io/gorm"
)

type ProductCategoryRepository interface {
	Create(ctx context.Context, category *entities.ProductCategory) error
	Update(ctx context.Context, category *entities.ProductCategory) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.ProductCategory, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsByCode(ctx context.Context, code string) (bool, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByNameExcludeID(ctx context.Context, name string, id uuid.UUID) (bool, error)
	ExistsByCodeExcludeID(ctx context.Context, code string, id uuid.UUID) (bool, error)
	ExistsBySlugExcludeID(ctx context.Context, slug string, id uuid.UUID) (bool, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	IsInUse(ctx context.Context, id uuid.UUID) (bool, error)
	FindAll(ctx context.Context, filter CategoryFilter) ([]entities.ProductCategory, int64, error)
}

type productCategoryRepository struct {
	db *gorm.DB
}

type CategoryFilter struct {
	Search *string
	Status *string
	Page   int
	Limit  int
}

func NewProductCategoryRepository(db *gorm.DB) ProductCategoryRepository {
	return &productCategoryRepository{db: db}
}

func (r *productCategoryRepository) Create(ctx context.Context, category *entities.ProductCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *productCategoryRepository) Update(ctx context.Context, category *entities.ProductCategory) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *productCategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.ProductCategory, error) {
	var category entities.ProductCategory
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
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

func (r *productCategoryRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.ProductCategory{}).
		Where("slug = ?", slug).
		Count(&count).Error
	return count > 0, err
}

func (r *productCategoryRepository) ExistsBySlugExcludeID(ctx context.Context, slug string, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.ProductCategory{}).
		Where("slug = ? AND id != ?", slug, id).
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

func (r *productCategoryRepository) ExistsByNameExcludeID(ctx context.Context, name string, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.ProductCategory{}).
		Where("name = ? AND id != ?", name, id).
		Count(&count).Error
	return count > 0, err
}

func (r *productCategoryRepository) ExistsByCodeExcludeID(ctx context.Context, code string, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.ProductCategory{}).
		Where("code = ? AND id != ?", code, id).
		Count(&count).Error
	return count > 0, err
}

// Implementation
func (r *productCategoryRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entities.ProductCategory{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *productCategoryRepository) IsInUse(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("products").
		Where("category_id = ? AND is_deleted = false", id).
		Count(&count).Error
	return count > 0, err
}

func (r *productCategoryRepository) FindAll(ctx context.Context, filter CategoryFilter) ([]entities.ProductCategory, int64, error) {
	var categories []entities.ProductCategory
	var total int64

	query := r.db.WithContext(ctx).Model(&entities.ProductCategory{})

	if filter.Search != nil {
		query = query.Where("name ILIKE ?", "%"+*filter.Search+"%")
	}
	if filter.Status != nil {
		isActive := *filter.Status == "ACTIVE"
		query = query.Where("is_active = ?", isActive)
	}

	query.Count(&total)

	offset := (filter.Page - 1) * filter.Limit
	err := query.
		Preload("Children").
		Offset(offset).
		Limit(filter.Limit).
		Find(&categories).Error

	return categories, total, err
}