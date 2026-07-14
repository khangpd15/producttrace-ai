package response

import (
	"time"

	"github.com/google/uuid"
)

// BatchDetailVariantResponse chứa thông tin variant đi kèm lô hàng
type BatchDetailVariantResponse struct {
	VariantID uuid.UUID `json:"variant_id"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	Barcode   *string   `json:"barcode"`
}

// BatchDetailProductResponse chứa thông tin sản phẩm cha của variant
type BatchDetailProductResponse struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
}

// BatchDetailResponse là DTO trả về đầy đủ thông tin chi tiết của một lô hàng
type BatchDetailResponse struct {
	ID               uuid.UUID                  `json:"id"`
	BatchCode        string                     `json:"batch_code"`
	ManufactureDate  *time.Time                 `json:"manufacture_date"`
	ExpiryDate       *time.Time                 `json:"expiry_date"`
	ImportedAt       *time.Time                 `json:"imported_at"`
	ManufacturerName *string                    `json:"manufacturer_name"`
	SupplierName     *string                    `json:"supplier_name"`
	OriginCountry    *string                    `json:"origin_country"`
	ProductionPlace  *string                    `json:"production_place"`
	Quantity         int                        `json:"quantity"`
	Status           string                     `json:"status"`
	CreatedAt        time.Time                  `json:"created_at"`
	UpdatedAt        time.Time                  `json:"updated_at"`
	Variant          BatchDetailVariantResponse `json:"variant"`
	Product          BatchDetailProductResponse `json:"product"`
}
