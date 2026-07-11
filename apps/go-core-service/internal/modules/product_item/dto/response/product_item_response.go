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

type VerifyQRRow struct {
	ItemCode         string     `json:"item_code"`
	SerialNumber     string     `json:"serial_number"`
	ItemStatus       string     `json:"item_status"`
	BatchCode        string     `json:"batch_code"`
	ManufactureDate  *time.Time `json:"manufacture_date"`
	ExpiryDate       *time.Time `json:"expiry_date"`
	ManufacturerName string     `json:"manufacturer_name"`
	SupplierName     string     `json:"supplier_name"`
	OriginCountry    string     `json:"origin_country"`
	ProductionPlace  string     `json:"production_place"`
	BatchStatus      string     `json:"batch_status"`
	ProductName      string     `json:"product_name"`
	VariantName      string     `json:"variant_name"`
	VariantSKU       string     `json:"variant_sku"`
}
