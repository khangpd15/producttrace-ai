package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	categoryEntities "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/entities"
	variantEntities "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/entities"
)

type Product struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	CategoryID   *uuid.UUID     `gorm:"type:uuid"`
	Name         string         `gorm:"type:varchar;not null"`
	Slug         *string        `gorm:"type:varchar"`
	Description  *string        `gorm:"type:text"`
	ThumbnailURL *string        `gorm:"type:text"`
	Tags         datatypes.JSON `gorm:"type:jsonb"`
	MetadataJSON datatypes.JSON `gorm:"type:jsonb"`
	Status       *string        `gorm:"type:varchar"`
	CreatedBy    *uuid.UUID     `gorm:"type:uuid"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	IsDeleted    bool `gorm:"default:false"`
	

	// Relations
	Category *categoryEntities.ProductCategory `gorm:"foreignKey:CategoryID"`
	Variants []variantEntities.ProductVariant  `gorm:"foreignKey:ProductID"`
}

func (Product) TableName() string {
	return "products"
}
