package entity

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"time"
)

type ProductItem struct {
	ID                     uuid.UUID  `gorm:"type:uuid;primaryKey"`
	VariantID              uuid.UUID  `gorm:"type:uuid;not null"`
	BatchID                *uuid.UUID `gorm:"type:uuid"`
	CurrentLocationPointID *uuid.UUID `gorm:"type:uuid"`
	ItemCode               string     `gorm:"type:varchar;not null"`
	SerialNumber           *string    `gorm:"type:varchar"`
	Status                 *string    `gorm:"type:varchar"`
	ProducedAt             *time.Time
	PackedAt               *time.Time
	SoldAt                 *time.Time
	RegisteredAt           *time.Time
	LastScannedAt          *time.Time
	MetadataJSON           datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt              time.Time
	UpdatedAt              time.Time
	IsDeleted              bool `gorm:"default:false"`

	// Relations
	Variant ProductVariant `gorm:"foreignKey:VariantID"`
}

func (ProductItem) TableName() string {
	return "product_items"
}
