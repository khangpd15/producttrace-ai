package request

type ListVariantRequest struct {
    Page   int     `form:"page,default=1" binding:"min=1"`
    Limit  int     `form:"limit,default=10" binding:"min=1,max=100"`
    Status *string `form:"status" binding:"omitempty,oneof=DRAFT ACTIVE INACTIVE OUT_OF_STOCK DISCONTINUED"`
}