package request

import "github.com/google/uuid"

type UpdateVariantRequest struct {
    ID       *uuid.UUID `json:"id"`        
    SKU      *string    `json:"sku"`
    Name     *string    `json:"name"`
    Barcode  *string    `json:"barcode"`
    Price    *float64   `json:"price" binding:"omitempty,min=0"`
    Currency *string    `json:"currency"`
    Images   []string   `json:"images"`
    Status   *string    `json:"status" binding:"omitempty,oneof=DRAFT ACTIVE INACTIVE OUT_OF_STOCK DISCONTINUED"`
}

type UpdateProductRequest struct {
    CategoryID   *uuid.UUID             `json:"category_id"`
    Name         *string                `json:"name"`
    Description  *string                `json:"description"`
    ThumbnailURL *string                `json:"thumbnail_url"`
    Status       *string                `json:"status" binding:"omitempty,oneof=DRAFT ACTIVE INACTIVE OUT_OF_STOCK DISCONTINUED"`
    Variants     []UpdateVariantRequest `json:"variants"`
}

