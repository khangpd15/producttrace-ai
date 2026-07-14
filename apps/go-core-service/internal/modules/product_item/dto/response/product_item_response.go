package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type ProductItemListItemDTO struct {
	ID          uuid.UUID `json:"id"`
	ItemCode    string    `json:"item_code"`
	ProductName string    `json:"product_name"`
	VariantName string    `json:"variant_name"`
	BatchCode   string    `json:"batch_code"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProductItemListResponse struct {
	Items []ProductItemListItemDTO `json:"items"`
	Meta  response.PaginationMeta  `json:"meta"`
}

type ProductItemEventDTO struct {
	EventName string    `json:"event_name"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

type ProductItemDetailDTO struct {
	ID                uuid.UUID             `json:"id"`
	ItemCode          string                `json:"item_code"`
	ProductName       string                `json:"product_name"`
	VariantName       string                `json:"variant_name"`
	BatchCode         string                `json:"batch_code"`
	Status            string                `json:"status"`
	VerificationToken string                `json:"verification_token"`
	CreatedAt         time.Time             `json:"created_at"`
	Events            []ProductItemEventDTO `json:"events"`
}
