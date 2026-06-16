package entity

import (
    "time"
    "github.com/google/uuid"
    "gorm.io/datatypes"
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
    IsDeleted    bool           `gorm:"default:false"`

    // Relations
    Category *ProductCategory  `gorm:"foreignKey:CategoryID"`
    Variants []ProductVariant  `gorm:"foreignKey:ProductID"`
}

func (Product) TableName() string {
    return "products"
}
