package entities

import (
    "time"
    "github.com/google/uuid"
)

type Attribute struct {
    ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
    CategoryID *uuid.UUID `gorm:"type:uuid"`
    Code       string     `gorm:"type:varchar;not null"`
    Label      string     `gorm:"type:varchar;not null"`
    CreatedAt  time.Time
    IsDeleted  bool       `gorm:"default:false"`

    // Relations
    Category *ProductCategory `gorm:"foreignKey:CategoryID"`
}

func (Attribute) TableName() string {
    return "attributes"
}