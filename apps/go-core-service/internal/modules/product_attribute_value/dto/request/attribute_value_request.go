package request

import "github.com/google/uuid"

type CreateAttributeValueRequest struct {
	AttributeID  uuid.UUID `json:"attribute_id" binding:"required"`
	Label        string    `json:"label" binding:"required"`
	ValueText    *string   `json:"value_text"`
	ValueNumber  *float64  `json:"value_number"`
	ValueBoolean *bool     `json:"value_boolean"`
}

type BulkCreateAttributeValuesRequest struct {
	Items []CreateAttributeValueRequest `json:"items" binding:"required,min=1,dive"`
}

type UpdateAttributeValueRequest struct {
	Label        *string  `json:"label"`
	ValueText    *string  `json:"value_text"`
	ValueNumber  *float64 `json:"value_number"`
	ValueBoolean *bool    `json:"value_boolean"`
}

type ListAttributeValueRequest struct {
	Page        int     `form:"page,default=1" binding:"min=1"`
	Limit       int     `form:"limit,default=10" binding:"min=1,max=100"`
	VariantID   *string `form:"variant_id"`
	AttributeID *string `form:"attribute_id"`
}
