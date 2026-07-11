package response

import (
	"time"

	"github.com/google/uuid"
)

type AttributeResponse struct {
	ID         uuid.UUID `json:"id"`
	CategoryID uuid.UUID `json:"category_id"`
	Code       string    `json:"code"`
	Label      string    `json:"label"`
	CreatedAt  time.Time `json:"created_at"`
}

type ListAttributeResponse struct {
	Data       []AttributeResponse `json:"data"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int                 `json:"total_pages"`
}
