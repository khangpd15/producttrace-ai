package repository

import (
	"context"
	"errors"
	"time"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/domain"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"gorm.io/gorm"
)

type locationRepository struct {
	db *gorm.DB
}

func NewLocationRepository(db *gorm.DB) LocationRepository {
	return &locationRepository{db: db}
}

func (r *locationRepository) Create(ctx context.Context, loc *domain.Location) error {
	result := r.db.WithContext(ctx).Create(loc)
	if result.Error != nil {
		return apperror.WrapDBError(result.Error, "location")
	}
	return nil
}

// GetByID truy vấn một Location theo ID (uuid), bỏ qua bản ghi đã soft-delete.
func (r *locationRepository) GetByID(ctx context.Context, id string) (*domain.Location, error) {
	var loc domain.Location
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&loc)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound("location")
		}
		return nil, apperror.WrapDBError(result.Error, "location")
	}
	return &loc, nil
}

// GetByCode truy vấn một Location theo code duy nhất, bỏ qua bản ghi đã soft-delete.
func (r *locationRepository) GetByCode(ctx context.Context, code string) (*domain.Location, error) {
	var loc domain.Location
	result := r.db.WithContext(ctx).
		Where("code = ?", code).
		First(&loc)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Code chưa tồn tại — trả nil, nil để caller kiểm tra existence
			return nil, nil
		}
		return nil, apperror.WrapDBError(result.Error, "location")
	}
	return &loc, nil
}

// ListAll trả về danh sách Location có hỗ trợ filter, tìm kiếm và phân trang.
// Các filter có thể để rỗng/nil để bỏ qua.
func (r *locationRepository) ListAll(
	ctx context.Context,
	city string,
	locType domain.LocationType,
	isActive *bool,
	keyword string,
	offset, limit int,
) ([]*domain.Location, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Location{})

	if city != "" {
		q = q.Where("city = ?", city)
	}
	if locType != "" {
		q = q.Where("type = ?", locType)
	}
	if isActive != nil {
		q = q.Where("is_active = ?", *isActive)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where(
			"code ILIKE ? OR name ILIKE ? OR address ILIKE ? OR ward ILIKE ? OR district ILIKE ? OR city ILIKE ? OR country ILIKE ?",
			like, like, like, like, like, like, like,
		)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperror.WrapDBError(err, "location")
	}

	var locs []*domain.Location
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&locs).Error; err != nil {
		return nil, 0, apperror.WrapDBError(err, "location")
	}

	return locs, total, nil
}

// Update cập nhật toàn bộ thông tin của một Location.
func (r *locationRepository) Update(ctx context.Context, loc *domain.Location) error {
	loc.UpdatedAt = time.Now()
	result := r.db.WithContext(ctx).
		Model(&domain.Location{}).
		Where("id = ?", loc.ID).
		Select("*").
		Updates(loc)
	if result.Error != nil {
		return apperror.WrapDBError(result.Error, "location")
	}
	if result.RowsAffected == 0 {
		return apperror.NewNotFound("location")
	}
	return nil
}

// HardDelete xóa vĩnh viễn một Location khỏi database bằng lệnh SQL DELETE.
func (r *locationRepository) HardDelete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).
		Exec("DELETE FROM locations WHERE id = ?", id)
	if result.Error != nil {
		return apperror.WrapDBError(result.Error, "location")
	}
	if result.RowsAffected == 0 {
		return apperror.NewNotFound("location")
	}
	return nil
}
