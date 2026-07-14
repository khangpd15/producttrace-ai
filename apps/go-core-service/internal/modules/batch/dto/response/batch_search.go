package response

import (
	"time"

	"github.com/google/uuid"
)

// SearchBatchItemDTO là DTO cho từng phần tử trong kết quả tìm kiếm lô hàng.
// JSON keys tuân thủ đúng spec UC-P2-BATCH-03 (camelCase).
type SearchBatchItemDTO struct {
	BatchID           uuid.UUID  `json:"batchId"`
	BatchCode         string     `json:"batchCode"`
	ProductName       string     `json:"productName"`
	ManufacturingDate *time.Time `json:"manufacturingDate"`
	Quantity          int        `json:"quantity"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"createdAt"`
}

// SearchBatchPaginationDTO chứa metadata phân trang theo spec UC-P2-BATCH-03.
type SearchBatchPaginationDTO struct {
	CurrentPage  int `json:"currentPage"`
	PageSize     int `json:"pageSize"`
	TotalRecords int `json:"totalRecords"`
	TotalPages   int `json:"totalPages"`
}

// SearchBatchResponse là DTO trả về cho endpoint GET /api/v1/batches/search.
// Cấu trúc này khác BatchListResponse (không có stats) và tuân thủ đúng
// output specification của UC-P2-BATCH-03.
type SearchBatchResponse struct {
	Items      []SearchBatchItemDTO     `json:"items"`
	Pagination SearchBatchPaginationDTO `json:"pagination"`
}
