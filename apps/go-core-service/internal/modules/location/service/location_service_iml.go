package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/domain"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/repository"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
)

type locationService struct {
	repo repository.LocationRepository
}

// NewLocationService khởi tạo một LocationService mới.
func NewLocationService(repo repository.LocationRepository) LocationService {
	return &locationService{repo: repo}
}

// ─── Helpers: mapping

// toResponse chuyển đổi domain.Location → dto.LocationResponse.
func toResponse(loc *domain.Location) *dto.LocationResponse {
	return &dto.LocationResponse{
		ID:               loc.ID,
		OwnerUserID:      loc.OwnerUserID,
		Code:             loc.Code,
		Name:             loc.Name,
		Type:             loc.Type,
		Phone:            loc.Phone,
		Email:            loc.Email,
		Address:          loc.Address,
		Ward:             loc.Ward,
		District:         loc.District,
		City:             loc.City,
		Country:          loc.Country,
		Latitude:         loc.Latitude,
		Longitude:        loc.Longitude,
		OpeningHoursJSON: loc.OpeningHoursJSON,
		IsActive:         loc.IsActive,
		CreatedAt:        loc.CreatedAt,
		UpdatedAt:        loc.UpdatedAt,
		GeoLocation: func() *dto.GeoLocationDTO {
			if loc.GeoLocation == nil {
				return nil
			}
			return &dto.GeoLocationDTO{
				Latitude:  loc.GeoLocation.Latitude,
				Longitude: loc.GeoLocation.Longitude,
			}
		}(),
	}
}

func (s *locationService) CreateLocation(ctx context.Context, req *dto.CreateLocationReq) (*dto.LocationResponse, error) {
	// Kiểm tra owner_user_id là UUID hợp lệ
	if _, err := uuid.Parse(req.OwnerUserID); err != nil {
		return nil, apperror.NewBadRequest("owner_user_id is not a valid UUID")
	}

	// Kiểm tra trùng code
	existing, err := s.repo.GetByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.NewConflict("location with code '" + req.Code + "' already exists")
	}

	var geoLoc *domain.GeoLocation
	if req.GeoLocation != nil {
		geoLoc = &domain.GeoLocation{
			Latitude:  req.GeoLocation.Latitude,
			Longitude: req.GeoLocation.Longitude,
		}
	} else if req.Latitude != nil && req.Longitude != nil {
		geoLoc = &domain.GeoLocation{
			Latitude:  *req.Latitude,
			Longitude: *req.Longitude,
		}
	}

	loc := &domain.Location{
		ID:               uuid.New().String(),
		OwnerUserID:      req.OwnerUserID,
		Code:             req.Code,
		Name:             req.Name,
		Type:             req.Type,
		Phone:            req.Phone,
		Email:            req.Email,
		Address:          req.Address,
		Ward:             req.Ward,
		District:         req.District,
		City:             req.City,
		Country:          "Vietnam",
		Latitude:         *req.Latitude,
		Longitude:        *req.Longitude,
		OpeningHoursJSON: req.OpeningHoursJSON,
		IsActive:         true,
		GeoLocation:      geoLoc,
	}

	if err := s.repo.Create(ctx, loc); err != nil {
		return nil, err
	}

	return toResponse(loc), nil
}

// GetLocationByID lấy thông tin Location theo ID.
func (s *locationService) GetLocationByID(ctx context.Context, id string) (*dto.LocationResponse, error) {
	// Validate UUID format trước khi query DB
	if _, err := uuid.Parse(id); err != nil {
		return nil, apperror.NewBadRequest("id is not a valid UUID")
	}

	loc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return toResponse(loc), nil
}

// ListLocations trả về danh sách Location theo filter và phân trang.
func (s *locationService) ListLocations(ctx context.Context, req *dto.ListLocationsReq) (*dto.ListLocationsResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	locs, total, err := s.repo.ListAll(ctx, req.City, req.Type, req.IsActive, offset, limit)
	if err != nil {
		return nil, err
	}

	data := make([]*dto.LocationResponse, 0, len(locs))
	for _, loc := range locs {
		data = append(data, toResponse(loc))
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &dto.ListLocationsResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetLocationByCode lấy thông tin Location theo code.
func (s *locationService) GetLocationByCode(ctx context.Context, code string) (*dto.LocationResponse, error) {
	if code == "" {
		return nil, apperror.NewBadRequest("code is required")
	}

	loc, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if loc == nil {
		return nil, apperror.NewNotFound("location")
	}

	return toResponse(loc), nil
}

// UpdateLocation cập nhật thông tin của một Location theo ID.
func (s *locationService) UpdateLocation(ctx context.Context, id string, req *dto.UpdateLocationReq) (*dto.LocationResponse, error) {
	// Validate UUID format trước khi query DB
	if _, err := uuid.Parse(id); err != nil {
		return nil, apperror.NewBadRequest("id is not a valid UUID")
	}

	// Kiểm tra tồn tại
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing.Name = req.Name
	existing.Type = req.Type
	existing.Phone = req.Phone
	existing.Email = req.Email
	existing.Address = req.Address
	existing.Ward = req.Ward
	existing.District = req.District
	existing.City = req.City
	existing.Latitude = *req.Latitude
	existing.Longitude = *req.Longitude
	existing.OpeningHoursJSON = req.OpeningHoursJSON

	if req.GeoLocation != nil {
		existing.GeoLocation = &domain.GeoLocation{
			Latitude:  req.GeoLocation.Latitude,
			Longitude: req.GeoLocation.Longitude,
		}
	} else if req.Latitude != nil && req.Longitude != nil {
		existing.GeoLocation = &domain.GeoLocation{
			Latitude:  *req.Latitude,
			Longitude: *req.Longitude,
		}
	}

	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return toResponse(existing), nil
}

// HardDeleteLocation xóa vĩnh viễn một Location khỏi database theo ID.
func (s *locationService) HardDeleteLocation(ctx context.Context, id string) error {
	// Validate UUID format trước khi query DB
	if _, err := uuid.Parse(id); err != nil {
		return apperror.NewBadRequest("id is not a valid UUID")
	}

	if err := s.repo.HardDelete(ctx, id); err != nil {
		return err
	}

	return nil
}
