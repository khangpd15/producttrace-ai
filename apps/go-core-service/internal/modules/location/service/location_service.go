package service

import (
	"context"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/dto"

)

// LocationService định nghĩa interface cho business logic.
type LocationService interface {
	CreateLocation(ctx context.Context, req *dto.CreateLocationReq) (*dto.LocationResponse, error)
	GetLocationByID(ctx context.Context, id string) (*dto.LocationResponse, error)
	GetLocationByCode(ctx context.Context, code string) (*dto.LocationResponse, error)
	ListLocations(ctx context.Context, req *dto.ListLocationsReq) (*dto.ListLocationsResponse, error)
	UpdateLocation(ctx context.Context, id string, req *dto.UpdateLocationReq) (*dto.LocationResponse, error)
	DeleteLocation(ctx context.Context, id string) error
	HardDeleteLocation(ctx context.Context, id string) error
	FindNearby(ctx context.Context, req *dto.FindNearbyReq) ([]*dto.LocationWithDistanceResponse, error)
}

