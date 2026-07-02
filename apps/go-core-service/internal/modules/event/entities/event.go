package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Event đại diện cho một sự kiện xảy ra trên một product item
// trong vòng đời truy xuất nguồn gốc (ví dụ: PRODUCED, PACKED, SHIPPED, SOLD, SCANNED).
type Event struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"            json:"id"`
	ProductItemID uuid.UUID      `gorm:"type:uuid;not null;index"        json:"product_item_id"`
	EventType     string         `gorm:"type:varchar(100);not null"      json:"event_type"`
	OccurredAt    time.Time      `gorm:"not null"                        json:"occurred_at"`
	Location      string         `gorm:"type:varchar(255)"               json:"location"`
	Actor         string         `gorm:"type:varchar(255)"               json:"actor"`
	Description   string         `gorm:"type:text"                       json:"description"`
	MetadataJSON  datatypes.JSON `gorm:"type:jsonb"                      json:"metadata,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"                  json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"                  json:"updated_at"`
	IsDeleted     bool           `gorm:"default:false"                   json:"is_deleted"`
}

func (Event) TableName() string {
	return "events"
}
