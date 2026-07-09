package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type BatchListItemDTO struct {
	ID              uuid.UUID  `json:"id"`
	BatchCode       string     `json:"batch_code"`
	VariantID       uuid.UUID  `json:"variant_id"`
	VariantName     string     `json:"variant_name"`
	Quantity        int        `json:"quantity"`
	ManufactureDate *time.Time `json:"manufacture_date"`
	ExpiryDate      *time.Time `json:"expiry_date"`
	OriginCountry   string     `json:"origin_country"`
	Status          string     `json:"status"`
}

type BatchStatsDTO struct {
	Total           int `json:"total"`
	Active          int `json:"active"`
	Expired         int `json:"expired"`
	RecalledBlocked int `json:"recalled_blocked"`
}

type BatchListResponse struct {
	Items []BatchListItemDTO      `json:"items"`
	Meta  response.PaginationMeta `json:"meta"`
	Stats BatchStatsDTO           `json:"stats"`
}
