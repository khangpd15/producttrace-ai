package request

// SearchBatchRequest chứa các tham số query cho API tìm kiếm lô hàng
// theo đặc tả UC-P2-BATCH-03.
//
// Quy tắc:
//   - keyword: tùy chọn, tối đa 100 ký tự. Khi có, hệ thống tìm ILIKE trên
//     batch_code, manufacturer_name, product_variant.name.
//   - page: mặc định 1, tối thiểu 1.
//   - pageSize: mặc định 10, tối thiểu 1, tối đa 100.
//   - sortBy: trường sắp xếp — "createdAt" (mặc định) hoặc "batchCode".
//   - sortOrder: chiều sắp xếp — "DESC" (mặc định) hoặc "ASC".
type SearchBatchRequest struct {
	Keyword   string `form:"keyword"             binding:"max=100"`
	Page      int    `form:"page,default=1"      binding:"min=1"`
	PageSize  int    `form:"pageSize,default=10" binding:"min=1,max=100"`
	SortBy    string `form:"sortBy,default=createdAt"`
	SortOrder string `form:"sortOrder,default=DESC"`
}
