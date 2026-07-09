package request

type UpdateVariantRequest struct {
    SKU      *string  `json:"sku"`
    Name     *string  `json:"name"`
    Barcode  *string  `json:"barcode"`
    Price    *float64 `json:"price" binding:"omitempty,min=0"`
    Currency *string  `json:"currency"`
    Images   []string `json:"images"`
    Status   *string  `json:"status" binding:"omitempty,oneof=DRAFT ACTIVE INACTIVE OUT_OF_STOCK DISCONTINUED"`
}