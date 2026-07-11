package entities

import (
	"time"

	"github.com/google/uuid"
)

type Attribute struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	CategoryID uuid.UUID `gorm:"type:uuid;column:category_id;not null"`
	Code       string    `gorm:"type:varchar;column:code;not null"`
	Label      string    `gorm:"type:varchar;column:label;not null"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	IsDeleted  bool      `gorm:"column:is_deleted;default:false"`
}

func (Attribute) TableName() string {
	return "attributes"
}
