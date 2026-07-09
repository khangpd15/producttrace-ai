package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/entities"
	"gorm.io/gorm"
)

type AttributeRepository interface {
	Create(ctx context.Context, attr *entities.Attribute) error
	Update(ctx context.Context, attr *entities.Attribute) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Attribute, error)
	FindAll(ctx context.Context, categoryID *uuid.UUID, search *string, offset, limit int) ([]entities.Attribute, int64, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ExistsByCodeAndCategory(ctx context.Context, code string, categoryID uuid.UUID) (bool, error)
	ExistsByCodeAndCategoryExcludeID(ctx context.Context, code string, categoryID uuid.UUID, excludeID uuid.UUID) (bool, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsInAttributeValues(ctx context.Context, id uuid.UUID) (bool, error)
}

type attributeRepository struct {
	db *gorm.DB
}

func NewAttributeRepository(db *gorm.DB) AttributeRepository {
	return &attributeRepository{db: db}
}

func (r *attributeRepository) Create(ctx context.Context, attr *entities.Attribute) error {
	return r.db.WithContext(ctx).Create(attr).Error
}

func (r *attributeRepository) Update(ctx context.Context, attr *entities.Attribute) error {
	return r.db.WithContext(ctx).Save(attr).Error
}

func (r *attributeRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Attribute, error) {
	var attr entities.Attribute
	err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = false", id).First(&attr).Error
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

func (r *attributeRepository) FindAll(ctx context.Context, categoryID *uuid.UUID, search *string, offset, limit int) ([]entities.Attribute, int64, error) {
	var attrs []entities.Attribute
	var total int64

	query := r.db.WithContext(ctx).Model(&entities.Attribute{}).Where("is_deleted = false")

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}
	if search != nil && *search != "" {
		query = query.Where("label ILIKE ? OR code ILIKE ?", "%"+*search+"%", "%"+*search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Order("created_at desc").Find(&attrs).Error
	return attrs, total, err
}

func (r *attributeRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entities.Attribute{}).Where("id = ?", id).Update("is_deleted", true).Error
}

func (r *attributeRepository) ExistsByCodeAndCategory(ctx context.Context, code string, categoryID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Attribute{}).
		Where("code = ? AND category_id = ? AND is_deleted = false", code, categoryID).
		Count(&count).Error
	return count > 0, err
}

func (r *attributeRepository) ExistsByCodeAndCategoryExcludeID(ctx context.Context, code string, categoryID uuid.UUID, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Attribute{}).
		Where("code = ? AND category_id = ? AND id != ? AND is_deleted = false", code, categoryID, excludeID).
		Count(&count).Error
	return count > 0, err
}

func (r *attributeRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Attribute{}).Where("id = ? AND is_deleted = false", id).Count(&count).Error
	return count > 0, err
}

func (r *attributeRepository) ExistsInAttributeValues(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("attribute_values").Where("attribute_id = ?", id).Count(&count).Error
	return count > 0, err
}
