package response

import (
	"github.com/google/uuid"
	attrValResponse "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/dto/response"
	"time"
)

type VariantResponse struct {
	ID         uuid.UUID                                `json:"id"`
	SKU        string                                    `json:"sku"`
	Name       string                                    `json:"name"`
	Barcode    *string                                   `json:"barcode"`
	Price      *float64                                  `json:"price"`
	Currency   *string                                   `json:"currency"`
	Images     []string                                  `json:"images"`
	Status     *string                                   `json:"status"`
	Attributes []attrValResponse.AttributeValueResponse `json:"attributes"`
	CreatedAt  time.Time                                 `json:"created_at"`
	UpdatedAt  time.Time                                 `json:"updated_at"`
}

type ProductResponse struct {
	ID           uuid.UUID              `json:"id"`
	CategoryID   *uuid.UUID             `json:"category_id"`
	Category     *string                `json:"category,omitempty"`
	Name         string                 `json:"name"`
	Slug         *string                `json:"slug"`
	Description  *string                `json:"description"`
	ThumbnailURL *string                `json:"thumbnail_url"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
	Status       *string                `json:"status"`
	CreatedBy    *uuid.UUID             `json:"created_by"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	Variants     []VariantResponse      `json:"variants"`
}