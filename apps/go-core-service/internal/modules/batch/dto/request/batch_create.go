package request

import (
	"time"

	"github.com/google/uuid"
)

// CreateBatchRequest là body JSON gửi lên khi tạo lô hàng mới.
// BatchCode sẽ được tự động sinh theo format [PREFIX]-[YEAR]-[SEQUENCE]
// nên client chỉ cần cung cấp Prefix (mã thương hiệu/nhà sản xuất).
//
// Ví dụ: prefix "APL" + năm 2026 → batch code "APL-2026-0001"
type CreateBatchRequest struct {
	VariantID        uuid.UUID  `json:"variant_id" binding:"required"`
	Prefix           string     `json:"prefix" binding:"required,min=2,max=20,alpha"`
	ManufactureDate  *time.Time `json:"manufacture_date"`
	ExpiryDate       *time.Time `json:"expiry_date"`
	ImportedAt       *time.Time `json:"imported_at"`
	ManufacturerName *string    `json:"manufacturer_name"`
	SupplierName     *string    `json:"supplier_name"`
	OriginCountry    *string    `json:"origin_country"`
	ProductionPlace  *string    `json:"production_place"`
	Quantity         int        `json:"quantity" binding:"min=0"`
}
