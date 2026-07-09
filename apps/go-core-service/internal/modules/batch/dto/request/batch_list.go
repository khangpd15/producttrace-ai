package request

type GetBatchListRequest struct {
	Page          int    `form:"page,default=1" binding:"min=1"`
	Limit         int    `form:"limit,default=10" binding:"min=1,max=100"`
	Search        string `form:"search"`
	Status        string `form:"status"`
	OriginCountry string `form:"origin_country"`
}
