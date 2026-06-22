package request

type ListProductRequest struct {
    Page       int     `form:"page,default=1" binding:"min=1"`
    Limit      int     `form:"limit,default=10" binding:"min=1,max=100"`
    Search     *string `form:"search"`
    CategoryID *string `form:"category_id"`
    Status     *string `form:"status" binding:"omitempty,oneof=DRAFT ACTIVE INACTIVE OUT_OF_STOCK DISCONTINUED"`
    SortBy     *string `form:"sort_by"`
    SortOrder  *string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}