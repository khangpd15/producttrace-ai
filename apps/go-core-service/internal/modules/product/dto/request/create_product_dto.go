package request

import (
	"github.com/google/uuid"
	attrValRequest "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/dto/request"
)

type CreateVariantRequest struct {
	SKU      string   `json:"sku" binding:"required"`
	Name     string   `json:"name" binding:"required"`
	Barcode  *string  `json:"barcode"`
	Price    *float64 `json:"price" binding:"omitempty,min=0"`
	Currency *string  `json:"currency"`
	Images   []string `json:"images"`
	Status   *string  `json:"status" binding:"omitempty,oneof=DRAFT ACTIVE INACTIVE OUT_OF_STOCK DISCONTINUED"`
	// Attributes: danh sách thuộc tính gắn kèm variant này ngay khi tạo product
	// (product -> variant -> attributes trong cùng 1 request/transaction).
	Attributes []attrValRequest.CreateAttributeValueRequest `json:"attributes" binding:"omitempty,dive"`
}

type CreateProductRequest struct {
	CategoryID   uuid.UUID              `json:"category_id" binding:"required"`
	Name         string                 `json:"name" binding:"required"`
	Slug         string                 `json:"slug" binding:"required"`
	Description  *string                `json:"description"`
	ThumbnailURL *string                `json:"thumbnail_url"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata"`
	Status       string                 `json:"status" binding:"required,oneof=DRAFT ACTIVE DISCONTINUED"`
	Variants     []CreateVariantRequest `json:"variants" binding:"required,min=1,dive"`
}