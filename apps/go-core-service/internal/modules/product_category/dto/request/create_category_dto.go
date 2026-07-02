package request

import "github.com/google/uuid"

type CreateCategoryRequest struct {
    Name        string     `json:"name" binding:"required,max=100"`
    Code        string     `json:"code" binding:"required"`
    ParentID    *uuid.UUID `json:"parent_id"`
    Description *string    `json:"description" binding:"omitempty,max=500"`
    Status      string     `json:"status" binding:"required,oneof=ACTIVE INACTIVE"`
}