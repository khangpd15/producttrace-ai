package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/domain"
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
		return fmt.Errorf("locationRepository.Create: %w", result.Error)
	}
	return nil
}

// GetByID truy vấn một Location theo ID (uuid), bỏ qua bản ghi đã soft-delete.
func (r *locationRepository) GetByID(ctx context.Context, id string) (*domain.Location, error) {
	var loc domain.Location
	result := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = false", id).
		First(&loc)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("locationRepository.GetByID: record not found for id=%s", id)
		}
		return nil, fmt.Errorf("locationRepository.GetByID: %w", result.Error)
	}
	return &loc, nil
}

// GetByCode truy vấn một Location theo code duy nhất, bỏ qua bản ghi đã soft-delete.
func (r *locationRepository) GetByCode(ctx context.Context, code string) (*domain.Location, error) {
	var loc domain.Location
	result := r.db.WithContext(ctx).
		Where("code = ? AND is_deleted = false", code).
		First(&loc)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("locationRepository.GetByCode: record not found for code=%s", code)
		}
		return nil, fmt.Errorf("locationRepository.GetByCode: %w", result.Error)
	}
	return &loc, nil
}

// ListAll trả về danh sách Location có hỗ trợ filter và phân trang.
// Các filter có thể để rỗng/nil để bỏ qua.
func (r *locationRepository) ListAll(
	ctx context.Context,
	city string,
	locType domain.LocationType,
	isActive *bool,
	offset, limit int,
) ([]*domain.Location, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Location{}).Where("is_deleted = false")

	if city != "" {
		q = q.Where("city = ?", city)
	}
	if locType != "" {
		q = q.Where("type = ?", locType)
	}
	if isActive != nil {
		q = q.Where("is_active = ?", *isActive)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("locationRepository.ListAll count: %w", err)
	}

	var locs []*domain.Location
	if err := q.Order("created_at DESC").Offset(offset).Limit(limit).Find(&locs).Error; err != nil {
		return nil, 0, fmt.Errorf("locationRepository.ListAll: %w", err)
	}

	return locs, total, nil
}

// Update cập nhật toàn bộ thông tin của một Location.
func (r *locationRepository) Update(ctx context.Context, loc *domain.Location) error {
	loc.UpdatedAt = time.Now()
	result := r.db.WithContext(ctx).
		Model(&domain.Location{}).
		Where("id = ? AND is_deleted = false", loc.ID).
		Updates(loc)
	if result.Error != nil {
		return fmt.Errorf("locationRepository.Update: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("locationRepository.Update: no rows affected for id=%s", loc.ID)
	}
	return nil
}

// Delete thực hiện Soft Delete: cập nhật is_deleted = true, is_active = false.
func (r *locationRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Location{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"is_active":  false,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("locationRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("locationRepository.Delete: record not found or already deleted for id=%s", id)
	}
	return nil
}

// HardDelete xóa vĩnh viễn một Location khỏi database bằng lệnh SQL DELETE.
func (r *locationRepository) HardDelete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).
		Exec("DELETE FROM locations WHERE id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("locationRepository.HardDelete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("locationRepository.HardDelete: record not found for id=%s", id)
	}
	return nil
}

// FindNearby tìm các Location trong bán kính radiusMeters (mét) từ tọa độ (lat, lng),
// sắp xếp theo khoảng cách tăng dần, sử dụng hàm PostGIS ST_DWithin và ST_Distance.
// Lưu ý PostGIS: ST_MakePoint(longitude, latitude) — longitude trước, latitude sau.
func (r *locationRepository) FindNearby(ctx context.Context, lat, lng float64, radiusMeters float64, limit int) ([]*domain.LocationWithDistance, error) {
	if limit <= 0 {
		limit = 20
	}

	// ST_DWithin trên geography tính theo đơn vị mét.
	// ST_Distance trên geography trả về khoảng cách mét.
	query := `
		SELECT
			l.*,
			ST_Distance(
				l.geo_location::geography,
				ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography
			) AS distance_meters
		FROM locations l
		WHERE
			l.is_deleted = false
			AND l.is_active  = true
			AND ST_DWithin(
				l.geo_location::geography,
				ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography,
				?
			)
		ORDER BY distance_meters ASC
		LIMIT ?
	`

	var results []*domain.LocationWithDistance
	// Thứ tự args: ST_Distance(lng, lat), ST_DWithin(lng, lat, radius), limit
	err := r.db.WithContext(ctx).
		Raw(query, lng, lat, lng, lat, radiusMeters, limit).
		Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("locationRepository.FindNearby: %w", err)
	}

	return results, nil
}
