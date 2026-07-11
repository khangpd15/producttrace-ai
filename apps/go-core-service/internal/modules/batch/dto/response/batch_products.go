package response

import (
	"time"

	"github.com/google/uuid"
)

// BatchProductLocationDTO chứa thông tin vị trí lưu kho hiện tại của sản phẩm.
// LEFT JOIN từ bảng locations qua current_location_id.
type BatchProductLocationDTO struct {
	LocationID uuid.UUID `json:"locationId"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
}

// BatchProductItemDTO là DTO cho từng sản phẩm đơn lẻ trong lô.
// Theo output spec UC-P2-BATCH-05 (camelCase JSON keys).
type BatchProductItemDTO struct {
	ItemID          uuid.UUID                `json:"itemId"`
	ItemCode        string                   `json:"itemCode"`
	SerialNumber    string                   `json:"serialNumber"`
	Status          string                   `json:"status"`
	CurrentLocation *BatchProductLocationDTO `json:"currentLocation"`
	CreatedAt       time.Time                `json:"createdAt"`
}

// BatchProductPaginationDTO theo đúng output spec UC-P2-BATCH-05.
type BatchProductPaginationDTO struct {
	CurrentPage  int `json:"currentPage"`
	PageSize     int `json:"pageSize"`
	TotalRecords int `json:"totalRecords"`
	TotalPages   int `json:"totalPages"`
}

// GetBatchProductsResponse là response DTO cho GET /api/v1/batches/:id/products.
type GetBatchProductsResponse struct {
	Items      []BatchProductItemDTO     `json:"items"`
	Pagination BatchProductPaginationDTO `json:"pagination"`
}
