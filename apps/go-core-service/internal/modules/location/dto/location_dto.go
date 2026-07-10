package dto

import (
	"time"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/domain"
)

type CreateLocationReq struct {
	OwnerUserID      string              `json:"ownerUserId" binding:"required"`
	Code             string              `json:"code" binding:"required"`
	Name             string              `json:"name" binding:"required"`
	Type             domain.LocationType `json:"type" binding:"required,oneof=WAREHOUSE STORE DEALER WARRANTY_CENTER"`
	Phone            string              `json:"phone"`
	Email            string              `json:"email" binding:"omitempty,email"`
	Address          string              `json:"address"`
	Ward             string              `json:"ward" binding:"required"`
	District         string              `json:"district" binding:"required"`
	City             string              `json:"city" binding:"required"`
	Latitude         *float64            `json:"latitude" binding:"required"`
	Longitude        *float64            `json:"longitude" binding:"required"`
	OpeningHoursJSON domain.OpeningHours `json:"openingHoursJson"`
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
	IsActive         *bool               `json:"isActive"`
	OpeningHoursJSON domain.OpeningHours `json:"openingHoursJson"`
}


// ListLocationsReq là query params cho GET /locations
type ListLocationsReq struct {
	Page    int    `form:"page"      binding:"omitempty,min=1"`
	Limit   int    `form:"limit"     binding:"omitempty,min=1,max=100"`
	City    string `form:"city"`
	Status  string `form:"status"    binding:"omitempty,oneof=ALL ACTIVE INACTIVE"`
	Type    string `form:"type"      binding:"omitempty,oneof=ALL WAREHOUSE STORE DEALER WARRANTY_CENTER"`
	Keyword string `form:"keyword"`
}

// ListLocationsResponse trả về danh sách có phân trang
type ListLocationsResponse struct {
	Data       []*LocationResponse `json:"data"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int                 `json:"totalPages"`
}

type LocationResponse struct {
	ID               string              `json:"id"`
	OwnerUserID      string              `json:"ownerUserId"`
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
	OpeningHoursJSON domain.OpeningHours `json:"openingHoursJson"`
	IsActive         bool                `json:"isActive"`
	CreatedAt        time.Time           `json:"createdAt"`
	UpdatedAt        time.Time           `json:"updatedAt"`
	GeoLocation      *GeoLocationDTO     `json:"geoLocation,omitempty"`
}

type GeoLocationDTO struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

