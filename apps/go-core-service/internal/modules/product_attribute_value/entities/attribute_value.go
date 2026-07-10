package entities

import (
	"time"

	"github.com/google/uuid"
)

type AttributeValue struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductVariantID uuid.UUID `gorm:"type:uuid;column:product_variant_id;not null"`
	AttributeID      uuid.UUID `gorm:"type:uuid;column:attribute_id;not null"`
	Label            string    `gorm:"type:varchar;column:label;not null"`
	ValueText        *string   `gorm:"column:value_text"`
	ValueNumber      *float64  `gorm:"column:value_number;type:decimal"`
	ValueBoolean     *bool     `gorm:"column:value_boolean"`
	CreatedAt        time.Time `gorm:"column:created_at"`
}

func (AttributeValue) TableName() string {
	return "attribute_values"
}
