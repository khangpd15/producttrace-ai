package request

// ImportBatchesRequest — DTO cho POST /batches/import (bulk import).
// Nhận các lô hàng đang vận chuyển (IN_TRANSIT) vào kho đích.
type ImportBatchesRequest struct {
	// BatchIDs: danh sách UUID của các batch cần nhập.
	BatchIDs []string `json:"batch_ids" binding:"required,min=1"`
	// Note: ghi chú tùy chọn từ người nhập kho.
	Note string `json:"note"`
}
