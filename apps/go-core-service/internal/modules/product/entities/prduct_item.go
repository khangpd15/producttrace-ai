package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/entities"
	"gorm.io/datatypes"
)

type ProductItem struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey"`
	VariantID         uuid.UUID  `gorm:"type:uuid;not null"`
	BatchID           *uuid.UUID `gorm:"type:uuid"`
	CurrentLocationID *uuid.UUID `gorm:"type:uuid"`
	ItemCode          string     `gorm:"type:varchar;not null"`
	SerialNumber      *string    `gorm:"type:varchar"`
	VerificationToken string     `gorm:"type:varchar"`
	Status            *string    `gorm:"type:varchar"`
	ProducedAt        *time.Time
	PackedAt          *time.Time
	SoldAt            *time.Time
	RegisteredAt      *time.Time
	LastScannedAt     *time.Time
	MetadataJSON      datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	IsDeleted         bool `gorm:"default:false"`

	// Relations
	Variant entities.ProductVariant `gorm:"foreignKey:VariantID"`
}

func (ProductItem) TableName() string {
	return "product_items"
}
