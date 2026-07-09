package request

// GetBatchListRequest là query params cho GET /api/v1/batches.
// ExcludeDraft không expose qua form — được service set theo role của user
// để thực thi BR-FIL-002: non-Admin không được xem lô DRAFT.
type GetBatchListRequest struct {
	Page          int    `form:"page,default=1" binding:"min=1"`
	Limit         int    `form:"limit,default=10" binding:"min=1,max=100"`
	Search        string `form:"search"`
	Status        string `form:"status"`
	OriginCountry string `form:"origin_country"`

	// ExcludeDraft được set bởi service (không parse từ query string).
	// Khi true, repository thêm điều kiện AND b.status != 'DRAFT'.
	ExcludeDraft bool `form:"-"`
}
