package response

import (
	"time"

	"github.com/google/uuid"
)

type AttributeValueResponse struct {
	ID               uuid.UUID `json:"id"`
	ProductVariantID uuid.UUID `json:"product_variant_id"`
	AttributeID      uuid.UUID `json:"attribute_id"`
	Label            string    `json:"label"`
	ValueText        *string   `json:"value_text,omitempty"`
	ValueNumber      *float64  `json:"value_number,omitempty"`
	ValueBoolean     *bool     `json:"value_boolean,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type ListAttributeValueResponse struct {
	Data       []AttributeValueResponse `json:"data"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	Limit      int                      `json:"limit"`
	TotalPages int                      `json:"total_pages"`
}
