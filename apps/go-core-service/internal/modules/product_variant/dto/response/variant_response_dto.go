package response

import (
    "time"
    "github.com/google/uuid"
)

type VariantResponse struct {
    ID        uuid.UUID `json:"id"`
    ProductID uuid.UUID `json:"product_id"`
    SKU       string    `json:"sku"`
    Name      string    `json:"name"`
    Barcode   *string   `json:"barcode"`
    Price     *float64  `json:"price"`
    Currency  *string   `json:"currency"`
    Images    []string  `json:"images"`
    Status    *string   `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type ListVariantResponse struct {
    Data  []VariantResponse `json:"data"`
    Total int64             `json:"total"`
}