package entities

import (
    "time"
    "github.com/google/uuid"
)

type ProductCategory struct {
    ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
    ParentID    *uuid.UUID `gorm:"type:uuid"`
    Code        *string    `gorm:"type:varchar"`
    Name        string     `gorm:"type:varchar;not null"`
    Slug        *string    `gorm:"type:varchar"`
    Description *string    `gorm:"type:text"`
    IconURL     *string    `gorm:"type:text"`
    IsActive    bool       `gorm:"default:true"`
    CreatedAt   time.Time
    UpdatedAt   time.Time

    // Relations
    Parent   *ProductCategory  `gorm:"foreignKey:ParentID"`
    Children []ProductCategory `gorm:"foreignKey:ParentID"`
}

func (ProductCategory) TableName() string {
    return "product_categories"
}