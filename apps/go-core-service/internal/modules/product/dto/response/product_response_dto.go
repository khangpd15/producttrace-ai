package response

import (
    "time"
    "github.com/google/uuid"
)

type VariantResponse struct {
    ID       uuid.UUID `json:"id"`
    SKU      string    `json:"sku"`
    Name     string    `json:"name"`
    Barcode  *string   `json:"barcode"`
    Price    *float64  `json:"price"`
    Currency *string   `json:"currency"`
    Images   []string  `json:"images"`
    Status   *string   `json:"status"`
}

type ProductResponse struct {
    ID           uuid.UUID         `json:"id"`
    CategoryID   *uuid.UUID        `json:"category_id"`
    Name         string            `json:"name"`
    Slug         *string           `json:"slug"`
    Description  *string           `json:"description"`
    ThumbnailURL *string           `json:"thumbnail_url"`
    Status       *string           `json:"status"`
    CreatedBy    *uuid.UUID        `json:"created_by"`
    CreatedAt    time.Time         `json:"created_at"`
    UpdatedAt    time.Time         `json:"updated_at"`
    Variants     []VariantResponse `json:"variants"`
}

type ListProductResponse struct {
    Data       []ProductResponse `json:"data"`
    Total      int64             `json:"total"`
    Page       int               `json:"page"`
    Limit      int               `json:"limit"`
    TotalPages int               `json:"total_pages"`
}