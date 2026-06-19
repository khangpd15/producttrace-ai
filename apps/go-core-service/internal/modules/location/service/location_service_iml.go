package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/domain"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/repository"
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
	}
}

// ─── Implementation ────────────────────────────────────────────────────────────

// CreateLocation validate và tạo mới một Location.
func (s *locationService) CreateLocation(ctx context.Context, req *dto.CreateLocationReq) (*dto.LocationResponse, error) {
	// Kiểm tra trùng code
	existing, _ := s.repo.GetByCode(ctx, req.Code)
	if existing != nil {
		return nil, fmt.Errorf("locationService.CreateLocation: code '%s' already exists", req.Code)
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
		IsDeleted:        false,
	}

	if err := s.repo.Create(ctx, loc); err != nil {
		return nil, fmt.Errorf("locationService.CreateLocation: %w", err)
	}

	return toResponse(loc), nil
}

// GetLocationByID lấy thông tin Location theo ID.
func (s *locationService) GetLocationByID(ctx context.Context, id string) (*dto.LocationResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("locationService.GetLocationByID: id is required")
	}

	loc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("locationService.GetLocationByID: %w", err)
	}

	return toResponse(loc), nil
}

// GetLocationByCode lấy thông tin Location theo code.
func (s *locationService) GetLocationByCode(ctx context.Context, code string) (*dto.LocationResponse, error) {
	if code == "" {
		return nil, fmt.Errorf("locationService.GetLocationByCode: code is required")
	}

	loc, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("locationService.GetLocationByCode: %w", err)
	}

	return toResponse(loc), nil
}

// UpdateLocation cập nhật thông tin của một Location theo ID.
func (s *locationService) UpdateLocation(ctx context.Context, id string, req *dto.UpdateLocationReq) (*dto.LocationResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("locationService.UpdateLocation: id is required")
	}

	// Kiểm tra tồn tại
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("locationService.UpdateLocation: %w", err)
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

	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("locationService.UpdateLocation: %w", err)
	}

	return toResponse(existing), nil
}

// DeleteLocation thực hiện soft-delete một Location theo ID.
func (s *locationService) DeleteLocation(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("locationService.DeleteLocation: id is required")
	}

	// Kiểm tra tồn tại trước khi xóa
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return fmt.Errorf("locationService.DeleteLocation: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("locationService.DeleteLocation: %w", err)
	}

	return nil
}

// FindNearby tìm các Location gần một tọa độ trong bán kính cho trước.
func (s *locationService) FindNearby(ctx context.Context, req *dto.FindNearbyReq) ([]*dto.LocationWithDistanceResponse, error) {
	if req.RadiusMeters <= 0 {
		return nil, fmt.Errorf("locationService.FindNearby: radius_meters must be greater than 0")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	results, err := s.repo.FindNearby(ctx, req.Latitude, req.Longitude, req.RadiusMeters, limit)
	if err != nil {
		return nil, fmt.Errorf("locationService.FindNearby: %w", err)
	}

	responses := make([]*dto.LocationWithDistanceResponse, 0, len(results))
	for _, r := range results {
		responses = append(responses, &dto.LocationWithDistanceResponse{
			LocationResponse: *toResponse(&r.Location),
			DistanceMeters:   r.DistanceMeters,
		})
	}

	return responses, nil
}
