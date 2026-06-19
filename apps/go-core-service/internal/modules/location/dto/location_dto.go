package dto

import (
	"time"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/domain"
)


type CreateLocationReq struct {
	OwnerUserID      string              `json:"owner_user_id" binding:"required"`
	Code             string              `json:"code" binding:"required"`
	Name             string              `json:"name" binding:"required"`
	Type             domain.LocationType `json:"type" binding:"required,oneof=WAREHOUSE STORE DEALER WARRANTY_CENTER"`
	Phone            string              `json:"phone"`
	Email            string              `json:"email" binding:"omitempty,email"`
	Address          string              `json:"address"`
	Ward             string              `json:"ward"`
	District         string              `json:"district"`
	City             string              `json:"city"`
	Latitude         *float64            `json:"latitude" binding:"required"`
	Longitude        *float64            `json:"longitude" binding:"required"`
	OpeningHoursJSON domain.OpeningHours `json:"opening_hours_json"`
}

type UpdateLocationReq struct {
	Name             string              `json:"name" binding:"required"`
	Type             domain.LocationType `json:"type" binding:"required,oneof=WAREHOUSE STORE DEALER WARRANTY_CENTER"`
	Phone            string              `json:"phone"`
	Email            string              `json:"email" binding:"omitempty,email"`
	Address          string              `json:"address"`
	Ward             string              `json:"ward"`
	District         string              `json:"district"`
	City             string              `json:"city"`
	Latitude         *float64            `json:"latitude" binding:"required"`
	Longitude        *float64            `json:"longitude" binding:"required"`
	IsActive         *bool               `json:"is_active"`
	OpeningHoursJSON domain.OpeningHours `json:"opening_hours_json"`
}

type FindNearbyReq struct {
	Latitude     float64 `json:"latitude"      binding:"required"`
	Longitude    float64 `json:"longitude"     binding:"required"`
	RadiusMeters float64 `json:"radius_meters" binding:"required,min=1,max=50000"`
	Limit        int     `json:"limit"         binding:"omitempty,min=1,max=100"`
}


type LocationResponse struct {
	ID               string              `json:"id"`
	OwnerUserID      string              `json:"owner_user_id"`
	Code             string              `json:"code"`
	Name             string              `json:"name"`
	Type             domain.LocationType `json:"type"`
	Phone            string              `json:"phone"`
	Email            string              `json:"email"`
	Address          string              `json:"address"`
	Ward             string              `json:"ward"`
	District         string              `json:"district"`
	City             string              `json:"city"`
	Country          string              `json:"country"`
	Latitude         float64             `json:"latitude"`
	Longitude        float64             `json:"longitude"`
	OpeningHoursJSON domain.OpeningHours `json:"opening_hours_json"`
	IsActive         bool                `json:"is_active"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

type LocationWithDistanceResponse struct {
	LocationResponse
	DistanceMeters float64 `json:"distance_meters"`
}
