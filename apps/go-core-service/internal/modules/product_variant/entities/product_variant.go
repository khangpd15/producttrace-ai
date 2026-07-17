package entities

import (
	"time"

	"github.com/google/uuid"
	attrValEntities "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/entities"
	"gorm.io/datatypes"
)

type ProductVariant struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ProductID  uuid.UUID      `gorm:"type:uuid;not null"`
	SKU        string         `gorm:"type:varchar;not null"`
	Name       string         `gorm:"type:varchar;not null"`
	Barcode    *string        `gorm:"type:varchar"`
	Price      *float64       `gorm:"type:decimal"`
	Currency   *string        `gorm:"type:varchar"`
	ImagesJSON datatypes.JSON `gorm:"type:jsonb"`
	Status     *string        `gorm:"type:varchar"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	IsDeleted  bool `gorm:"default:false"`

	AttributeValues []attrValEntities.AttributeValue `gorm:"-"`
}

func (ProductVariant) TableName() string {
	return "product_variants"
}
