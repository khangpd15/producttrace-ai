package request

// GetBatchProductsRequest chứa query params cho API xem sản phẩm trong lô.
// Theo UC-P2-BATCH-05: pagination bắt buộc (BR-BPR-001).
type GetBatchProductsRequest struct {
	Page    int    `form:"page,default=1"    binding:"min=1"`
	Limit   int    `form:"limit,default=20"  binding:"min=1,max=100"`
	Status  string `form:"status"`
	Keyword string `form:"keyword"           binding:"max=50"`
}
