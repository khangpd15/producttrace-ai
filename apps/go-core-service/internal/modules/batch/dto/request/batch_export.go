package request

// ExportBatchRequest — legacy DTO cho POST /batches/:id/export (single batch).
// Giữ lại để backward compatibility.
type ExportBatchRequest struct {
	DestinationLocation string `json:"destination_location" binding:"required"`
	Quantity            int    `json:"quantity" binding:"required,min=1"`
	OperatorName        string `json:"operator_name" binding:"required"`
	Notes               string `json:"notes"`
}

// ExportBatchesRequest — DTO cho POST /batches/export (bulk export).
// Xuất toàn bộ ProductItems của tất cả batch đã chọn.
// Không còn quantity và operator_name — backend tự lấy từ JWT context.
type ExportBatchesRequest struct {
	// BatchIDs: danh sách UUID của các batch cần xuất (ít nhất 1 phần tử).
	BatchIDs []string `json:"batch_ids" binding:"required,min=1"`
	// DestinationLocationID: UUID của Location đích (phải tồn tại trong bảng locations).
	DestinationLocationID string `json:"destination_location_id" binding:"required"`
	// Note: ghi chú tùy chọn, ví dụ: mục đích xuất kho.
	Note string `json:"note"`
}

