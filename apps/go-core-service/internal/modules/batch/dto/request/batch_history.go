package request

// GetBatchHistoryRequest chứa các query params phân trang cho API lịch sử lô hàng.
// Theo UC-P2-BATCH-06: page mặc định 1, limit mặc định 15.
type GetBatchHistoryRequest struct {
	Page  int `form:"page,default=1"   binding:"min=1"`
	Limit int `form:"limit,default=15" binding:"min=1,max=100"`
}
