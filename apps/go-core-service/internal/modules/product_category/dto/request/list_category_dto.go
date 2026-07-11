package request

type ListCategoryRequest struct {
	Search *string `form:"search"`
	Status *string `form:"status" binding:"omitempty,oneof=ACTIVE INACTIVE"`
	Page   int     `form:"page,default=1" binding:"min=1"`
	Limit  int     `form:"limit,default=10" binding:"min=1,max=100"`
}
