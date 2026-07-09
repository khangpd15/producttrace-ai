package request

// UpdateBatchStatusRequest là body JSON khi cập nhật status của Batch.
// Chỉ cho phép thay đổi duy nhất field status.
type UpdateBatchStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
