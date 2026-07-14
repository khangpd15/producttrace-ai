package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/dbctx"
	"gorm.io/gorm"
)

// InjectTx / GetDB giờ dùng chung từ package pkg/dbctx (xem dbctx.go)
// để đảm bảo product -> product_variant -> attribute_value cascade
// create/delete nằm trong đúng 1 transaction.

type AttributeValueRepository interface {
	Create(ctx context.Context, val *entities.AttributeValue) error
	CreateBulk(ctx context.Context, vals []entities.AttributeValue) error
	Update(ctx context.Context, val *entities.AttributeValue) error
	Delete(ctx context.Context, id uuid.UUID) error
	// DeleteByVariantID xoá cứng toàn bộ attribute value thuộc 1 variant
	// (dùng khi xoá riêng lẻ 1 variant, hoặc khi cascade từ product xuống).
	DeleteByVariantID(ctx context.Context, variantID uuid.UUID) error
	// DeleteByProductID xoá cứng toàn bộ attribute value thuộc tất cả
	// variant của 1 product (dùng khi xoá product).
	DeleteByProductID(ctx context.Context, productID uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.AttributeValue, error)
	FindByVariantID(ctx context.Context, variantID uuid.UUID) ([]entities.AttributeValue, error)
	FindAll(ctx context.Context, variantID *uuid.UUID, attributeID *uuid.UUID, offset, limit int) ([]entities.AttributeValue, int64, error)
	ExistsByVariantAndAttribute(ctx context.Context, variantID uuid.UUID, attributeID uuid.UUID) (bool, error)
	ExistsByVariantAndAttributeExcludeID(ctx context.Context, variantID uuid.UUID, attributeID uuid.UUID, excludeID uuid.UUID) (bool, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}

type attributeValueRepository struct {
	db *gorm.DB
}

func NewAttributeValueRepository(db *gorm.DB) AttributeValueRepository {
	return &attributeValueRepository{db: db}
}

func (r *attributeValueRepository) Create(ctx context.Context, val *entities.AttributeValue) error {
	return dbctx.GetDB(ctx, r.db).Create(val).Error
}

func (r *attributeValueRepository) CreateBulk(ctx context.Context, vals []entities.AttributeValue) error {
	return dbctx.GetDB(ctx, r.db).Create(&vals).Error
}

func (r *attributeValueRepository) Update(ctx context.Context, val *entities.AttributeValue) error {
	return dbctx.GetDB(ctx, r.db).Save(val).Error
}

func (r *attributeValueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return dbctx.GetDB(ctx, r.db).Where("id = ?", id).Delete(&entities.AttributeValue{}).Error
}

func (r *attributeValueRepository) DeleteByVariantID(ctx context.Context, variantID uuid.UUID) error {
	return dbctx.GetDB(ctx, r.db).
		Where("product_variant_id = ?", variantID).
		Delete(&entities.AttributeValue{}).Error
}

func (r *attributeValueRepository) DeleteByProductID(ctx context.Context, productID uuid.UUID) error {
	return dbctx.GetDB(ctx, r.db).
		Where("product_variant_id IN (SELECT id FROM product_variants WHERE product_id = ?)", productID).
		Delete(&entities.AttributeValue{}).Error
}

func (r *attributeValueRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.AttributeValue, error) {
	var val entities.AttributeValue
	err := dbctx.GetDB(ctx, r.db).Where("id = ?", id).First(&val).Error
	if err != nil {
		return nil, err
	}
	return &val, nil
}

func (r *attributeValueRepository) FindByVariantID(ctx context.Context, variantID uuid.UUID) ([]entities.AttributeValue, error) {
	var vals []entities.AttributeValue
	err := dbctx.GetDB(ctx, r.db).Where("product_variant_id = ?", variantID).Order("created_at asc").Find(&vals).Error
	return vals, err
}

func (r *attributeValueRepository) FindAll(ctx context.Context, variantID *uuid.UUID, attributeID *uuid.UUID, offset, limit int) ([]entities.AttributeValue, int64, error) {
	var vals []entities.AttributeValue
	var total int64

	query := dbctx.GetDB(ctx, r.db).Model(&entities.AttributeValue{})

	if variantID != nil {
		query = query.Where("product_variant_id = ?", *variantID)
	}
	if attributeID != nil {
		query = query.Where("attribute_id = ?", *attributeID)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Order("created_at desc").Find(&vals).Error
	return vals, total, err
}

func (r *attributeValueRepository) ExistsByVariantAndAttribute(ctx context.Context, variantID uuid.UUID, attributeID uuid.UUID) (bool, error) {
	var count int64
	err := dbctx.GetDB(ctx, r.db).Model(&entities.AttributeValue{}).
		Where("product_variant_id = ? AND attribute_id = ?", variantID, attributeID).
		Count(&count).Error
	return count > 0, err
}

func (r *attributeValueRepository) ExistsByVariantAndAttributeExcludeID(ctx context.Context, variantID uuid.UUID, attributeID uuid.UUID, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := dbctx.GetDB(ctx, r.db).Model(&entities.AttributeValue{}).
		Where("product_variant_id = ? AND attribute_id = ? AND id != ?", variantID, attributeID, excludeID).
		Count(&count).Error
	return count > 0, err
}

func (r *attributeValueRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := dbctx.GetDB(ctx, r.db).Model(&entities.AttributeValue{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
