package request

import "github.com/google/uuid"

type UpdateCategoryRequest struct {
	Name        *string    `json:"name" binding:"omitempty,max=100"`
	Code        *string    `json:"code"`
	ParentID    *uuid.UUID `json:"parent_id"`
	Description *string    `json:"description" binding:"omitempty,max=500"`
	Status      *string    `json:"status" binding:"omitempty,oneof=ACTIVE INACTIVE"`
}
