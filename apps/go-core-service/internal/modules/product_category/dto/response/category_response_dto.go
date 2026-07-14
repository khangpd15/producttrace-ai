package response

import (
	"github.com/google/uuid"
	"time"
)

type CategoryResponse struct {
	ID          uuid.UUID          `json:"id"`
	Name        string             `json:"name"`
	Code        *string            `json:"code"`
	Slug        *string            `json:"slug"`
	ParentID    *uuid.UUID         `json:"parent_id"`
	Description *string            `json:"description"`
	Icon        *string            `json:"icon"`
	IsActive    bool               `json:"is_active"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	Children    []CategoryResponse `json:"children,omitempty"`
}

type ListCategoryResponse struct {
	Data       []CategoryResponse `json:"data"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}