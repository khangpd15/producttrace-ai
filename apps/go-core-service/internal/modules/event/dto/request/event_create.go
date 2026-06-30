package request

import (
	"time"

	"github.com/google/uuid"
)

// CreateEventRequest là body JSON gửi lên khi tạo một event mới cho product item.
//
// EventType là loại sự kiện, ví dụ: PRODUCED, PACKED, SHIPPED, SOLD, SCANNED.
// OccurredAt mặc định là thời điểm hiện tại nếu không cung cấp.
type CreateEventRequest struct {
	ProductItemID uuid.UUID  `json:"product_item_id" binding:"required"`
	EventType     string     `json:"event_type"      binding:"required,min=1,max=100"`
	OccurredAt    *time.Time `json:"occurred_at"`
	Location      string     `json:"location"        binding:"max=255"`
	Actor         string     `json:"actor"           binding:"max=255"`
	Description   string     `json:"description"`
	Metadata      any        `json:"metadata"`
}
