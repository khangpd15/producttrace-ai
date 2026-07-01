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
	GeoLocation      *GeoLocationDTO     `json:"geo_location,omitempty"`
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
	GeoLocation      *GeoLocationDTO     `json:"geo_location,omitempty"`
}


// ListLocationsReq là query params cho GET /locations
type ListLocationsReq struct {
	Page     int                 `form:"page"      binding:"omitempty,min=1"`
	Limit    int                 `form:"limit"     binding:"omitempty,min=1,max=100"`
	City     string              `form:"city"`
	Type     domain.LocationType `form:"type"      binding:"omitempty,oneof=WAREHOUSE STORE DEALER WARRANTY_CENTER"`
	IsActive *bool               `form:"is_active"`
}

// ListLocationsResponse trả về danh sách có phân trang
type ListLocationsResponse struct {
	Data       []*LocationResponse `json:"data"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int                 `json:"total_pages"`
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
	GeoLocation      *GeoLocationDTO     `json:"geo_location,omitempty"`
}

type GeoLocationDTO struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

