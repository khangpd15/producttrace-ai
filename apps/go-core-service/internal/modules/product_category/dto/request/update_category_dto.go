package request

import "github.com/google/uuid"

type UpdateCategoryRequest struct {
	Name        *string    `json:"name"`
	Code        *string    `json:"code"`
	Slug        *string    `json:"slug"`
	ParentID    *uuid.UUID `json:"parent_id"`
	Description *string    `json:"description"`
	Icon        *string    `json:"icon"`
	Status      *string    `json:"status" binding:"omitempty,oneof=ACTIVE INACTIVE"`
}
