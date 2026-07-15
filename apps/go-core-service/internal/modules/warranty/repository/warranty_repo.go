package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/entity"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"gorm.io/gorm"
)

type WarrantyRepository interface {
	Create(ctx context.Context, w *entity.Warranty) error
	Update(ctx context.Context, w *entity.Warranty) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Warranty, error)
	FindByProductItemID(ctx context.Context, itemID uuid.UUID) (*entity.Warranty, error)
	FindByWarrantyCode(ctx context.Context, code string) (*entity.Warranty, error)
	FindAll(ctx context.Context, search string, status string, ownerID *uuid.UUID, page, limit int) ([]*entity.Warranty, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetStats(ctx context.Context) (map[string]int64, error)
	UpdateExpiredStatus(ctx context.Context) (int64, error)
}

type gormWarrantyRepository struct {
	db *gorm.DB
}

func NewWarrantyRepository(db *gorm.DB) WarrantyRepository {
	return &gormWarrantyRepository{db: db}
}

func (r *gormWarrantyRepository) Create(ctx context.Context, w *entity.Warranty) error {
	return r.db.WithContext(ctx).Create(w).Error
}

func (r *gormWarrantyRepository) Update(ctx context.Context, w *entity.Warranty) error {
	return r.db.WithContext(ctx).Save(w).Error
}

func (r *gormWarrantyRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Warranty, error) {
	var w entity.Warranty
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&w).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperror.WrapDBError(err, "warranties")
	}
	return &w, nil
}

func (r *gormWarrantyRepository) FindByProductItemID(ctx context.Context, itemID uuid.UUID) (*entity.Warranty, error) {
	var w entity.Warranty
	err := r.db.WithContext(ctx).Where("product_item_id = ?", itemID).Order("created_at DESC").First(&w).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperror.WrapDBError(err, "warranties")
	}
	return &w, nil
}

func (r *gormWarrantyRepository) FindByWarrantyCode(ctx context.Context, code string) (*entity.Warranty, error) {
	var w entity.Warranty
	err := r.db.WithContext(ctx).Where("warranty_code = ?", code).First(&w).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperror.WrapDBError(err, "warranties")
	}
	return &w, nil
}

func (r *gormWarrantyRepository) FindAll(ctx context.Context, search string, status string, ownerID *uuid.UUID, page, limit int) ([]*entity.Warranty, int64, error) {
	query := r.db.WithContext(ctx).Table("warranties w").
		Select("w.*, pi.item_code as item_code, pi.serial_number as serial_number, u.full_name as owner_name, u.email as owner_email").
		Joins("LEFT JOIN product_items pi ON w.product_item_id = pi.id").
		Joins("LEFT JOIN users u ON w.owner_id = u.id")

	if ownerID != nil {
		query = query.Where("w.owner_id = ?", *ownerID)
	}

	if status != "" && status != "ALL" {
		query = query.Where("w.status = ?", status)
	}

	if search != "" {
		s := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(w.warranty_code) LIKE ? OR LOWER(pi.item_code) LIKE ? OR LOWER(pi.serial_number) LIKE ? OR LOWER(u.full_name) LIKE ? OR LOWER(u.email) LIKE ?", s, s, s, s, s)
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, apperror.WrapDBError(err, "warranties")
	}

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	var results []*entity.Warranty
	err = query.Order("w.created_at DESC").Limit(limit).Offset(offset).Scan(&results).Error
	if err != nil {
		return nil, 0, apperror.WrapDBError(err, "warranties")
	}

	return results, count, nil
}

func (r *gormWarrantyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Warranty{}).Error
}

func (r *gormWarrantyRepository) GetStats(ctx context.Context) (map[string]int64, error) {
	var results []struct {
		Status string
		Count  int64
	}

	err := r.db.WithContext(ctx).Model(&entity.Warranty{}).Select("status, count(*) as count").Group("status").Scan(&results).Error
	if err != nil {
		return nil, apperror.WrapDBError(err, "warranties")
	}

	stats := map[string]int64{
		"total":     0,
		"active":    0,
		"inactive":  0,
		"expired":   0,
		"claimed":   0,
		"resolved":  0,
		"rejected":  0,
		"cancelled": 0,
	}

	for _, item := range results {
		key := strings.ToLower(item.Status)
		stats[key] = item.Count
		stats["total"] += item.Count
	}

	return stats, nil
}

func (r *gormWarrantyRepository) UpdateExpiredStatus(ctx context.Context) (int64, error) {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&entity.Warranty{}).
		Where("status = ? AND end_date IS NOT NULL AND end_date < ?", entity.WarrantyStatusActive, now).
		Update("status", entity.WarrantyStatusExpired)
	
	if result.Error != nil {
		return 0, apperror.WrapDBError(result.Error, "warranties")
	}
	return result.RowsAffected, nil
}
