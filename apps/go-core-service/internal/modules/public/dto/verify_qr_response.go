package dto

import "time"

// VerifyQRBatchInfo thông tin lô hàng trả về khi scan QR.
type VerifyQRBatchInfo struct {
	BatchCode        string     `json:"batch_code"`
	ManufactureDate  *time.Time `json:"manufacture_date"`
	ExpiryDate       *time.Time `json:"expiry_date"`
	ManufacturerName string     `json:"manufacturer_name"`
	SupplierName     string     `json:"supplier_name"`
	OriginCountry    string     `json:"origin_country"`
	ProductionPlace  string     `json:"production_place"`
	Status           string     `json:"batch_status"`
}

// VerifyQRProductInfo thông tin sản phẩm trả về khi scan QR.
type VerifyQRProductInfo struct {
	ProductName string `json:"product_name"`
	VariantName string `json:"variant_name"`
	VariantSKU  string `json:"variant_sku"`
}

// VerifyQRResponse là DTO trả về sau khi xác thực QR code thành công.
type VerifyQRResponse struct {
	ItemCode     string              `json:"item_code"`
	SerialNumber string              `json:"serial_number"`
	Status       string              `json:"item_status"`
	ScannedAt    time.Time           `json:"scanned_at"`
	Batch        VerifyQRBatchInfo   `json:"batch"`
	Product      VerifyQRProductInfo `json:"product"`
}
