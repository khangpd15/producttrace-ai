package dto

import "github.com/google/uuid"

type CreateVariantRequest struct {
    SKU      string   `json:"sku" binding:"required"`
    Name     string   `json:"name" binding:"required"`
    Barcode  *string  `json:"barcode"`
    Price    *float64 `json:"price" binding:"omitempty,min=0"`
    Currency *string  `json:"currency"`
    Images   []string `json:"images"`
}

type CreateProductRequest struct {
    CategoryID   uuid.UUID              `json:"category_id" binding:"required"`
    Name         string                 `json:"name" binding:"required"`
    Description  *string                `json:"description"`
    ThumbnailURL *string                `json:"thumbnail_url"`
    Status       string                 `json:"status" binding:"required,oneof=DRAFT ACTIVE DISCONTINUED"`
    Variants     []CreateVariantRequest `json:"variants" binding:"required,min=1,dive"`
}