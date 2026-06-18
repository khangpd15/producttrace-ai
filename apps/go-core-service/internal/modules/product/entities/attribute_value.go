package entities

import (
	"time"

	"github.com/google/uuid"
)

type AttributeValue struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductVariantID uuid.UUID `gorm:"type:uuid;not null"`
	AttributeID      uuid.UUID `gorm:"type:uuid;not null"`
	Label            string    `gorm:"type:varchar;not null"`
	ValueText        *string   `gorm:"type:text"`
	ValueNumber      *float64  `gorm:"type:decimal"`
	ValueBoolean     *bool
	CreatedAt        time.Time

	// Relations
	Attribute      Attribute      `gorm:"foreignKey:AttributeID"`
	ProductVariant ProductVariant `gorm:"foreignKey:ProductVariantID"`
}

func (AttributeValue) TableName() string {
	return "attribute_values"
}
