package request

import "github.com/google/uuid"

type GetProductItemListRequest struct {
	BatchID *uuid.UUID `form:"batch_id"`
	Status  *string    `form:"status"`
	Page    int        `form:"page,default=1"`
	Limit   int        `form:"limit,default=10"`
}
