package request

import "github.com/google/uuid"

type CreateCategoryRequest struct {
	Name        string     `json:"name" binding:"required"`
	Code        string     `json:"code" binding:"required"`
	Slug        *string    `json:"slug"`
	ParentID    *uuid.UUID `json:"parent_id"`
	Description *string    `json:"description"`
	Icon        *string    `json:"icon"`
	Status      string     `json:"status" binding:"required,oneof=ACTIVE INACTIVE"`
}
