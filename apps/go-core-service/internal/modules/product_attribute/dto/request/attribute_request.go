package request

import "github.com/google/uuid"

type CreateAttributeRequest struct {
	CategoryID uuid.UUID `json:"category_id" binding:"required"`
	Code       string    `json:"code" binding:"required"`
	Label      string    `json:"label" binding:"required"`
}

type UpdateAttributeRequest struct {
	Code  *string `json:"code"`
	Label *string `json:"label"`
}

type ListAttributeRequest struct {
	Page       int     `form:"page,default=1" binding:"min=1"`
	Limit      int     `form:"limit,default=10" binding:"min=1,max=100"`
	Search     *string `form:"search"`
	CategoryID *string `form:"category_id"`
}
