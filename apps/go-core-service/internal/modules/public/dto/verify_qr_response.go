package dto

import (
	"time"

	"github.com/google/uuid"
)

// ─── Sub-DTOs ────────────────────────────────────────────────────────────────

// VerifyQRProductInfo thông tin sản phẩm trả về khi scan QR.
type VerifyQRProductInfo struct {
	ProductName  string `json:"productName"`
	Description  string `json:"description"`
	ThumbnailURL string `json:"thumbnailUrl"`
	CategoryName string `json:"categoryName"`
	VariantName  string `json:"variantName"`
	VariantSKU   string `json:"variantSku"`
	Barcode      string `json:"barcode"`
}

// VerifyQRBatchInfo thông tin lô hàng trả về khi scan QR.
type VerifyQRBatchInfo struct {
	BatchCode        string     `json:"batchCode"`
	ManufactureDate  *time.Time `json:"manufactureDate"`
	ExpiryDate       *time.Time `json:"expiryDate"`
	ManufacturerName string     `json:"manufacturerName"`
	SupplierName     string     `json:"supplierName"`
	OriginCountry    string     `json:"originCountry"`
	ProductionPlace  string     `json:"productionPlace"`
	Status           string     `json:"batchStatus"`
}

// VerifyQROwnership thông tin chủ sở hữu hiện tại.
type VerifyQROwnership struct {
	OwnerName     string    `json:"ownerName"`
	RegisteredAt  time.Time `json:"registeredAt"`
	OwnershipType string    `json:"ownershipType"`
	Status        string    `json:"status"`
}

// VerifyQRWarranty thông tin bảo hành.
type VerifyQRWarranty struct {
	ClaimNumber string    `json:"claimNumber"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
}

// VerifyQRLocation vị trí hiện tại của sản phẩm.
type VerifyQRLocation struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Address string `json:"address"`
	City    string `json:"city"`
}

// VerifyQREvent một sự kiện trong lịch sử truy vết.
type VerifyQREvent struct {
	EventType   string    `json:"eventType"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	ActorName   string    `json:"actorName"`
	OccurredAt  time.Time `json:"occurredAt"`
}

// ─── Main Response ───────────────────────────────────────────────────────────

// VerifyQRResponse là DTO trả về sau khi xác thực QR code thành công.
type VerifyQRResponse struct {
	ProductItemID   uuid.UUID           `json:"productItemId"`
	ProductID       *uuid.UUID          `json:"productId"`
	VariantID       *uuid.UUID          `json:"variantId"`
	BatchID         *uuid.UUID          `json:"batchId"`
	ItemCode        string              `json:"itemCode"`
	SerialNumber    string              `json:"serialNumber"`
	Status          string              `json:"itemStatus"`
	ScannedAt       time.Time           `json:"scannedAt"`
	OwnershipStatus string              `json:"ownershipStatus,omitempty"`
	WarrantyStatus  string              `json:"warrantyStatus,omitempty"`
	Product         VerifyQRProductInfo `json:"product"`
	Batch           VerifyQRBatchInfo   `json:"batch"`
	Ownership       *VerifyQROwnership  `json:"ownership"`     // null nếu chưa đăng ký
	Warranty        *VerifyQRWarranty   `json:"warranty"`      // null nếu không có
	Location        *VerifyQRLocation   `json:"location"`      // null nếu chưa có
	TraceHistory    []VerifyQREvent     `json:"traceHistory"`  // empty array nếu chưa có
}
