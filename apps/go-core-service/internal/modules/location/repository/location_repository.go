package repository

import (
	"context"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/domain"
	
)

// LocationRepository định nghĩa interface cho tầng data access.
type LocationRepository interface {
	Create(ctx context.Context, loc *domain.Location) error
	GetByID(ctx context.Context, id string) (*domain.Location, error)
	GetByCode(ctx context.Context, code string) (*domain.Location, error)
	ListAll(ctx context.Context, city string, locType domain.LocationType, isActive *bool, offset, limit int) ([]*domain.Location, int64, error)
	Update(ctx context.Context, loc *domain.Location) error
	Delete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
	FindNearby(ctx context.Context, lat, lng float64, radiusMeters float64, limit int) ([]*domain.LocationWithDistance, error)
}

