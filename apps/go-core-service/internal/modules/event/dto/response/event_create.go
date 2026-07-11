package response

import (
	"time"

	"github.com/google/uuid"
)

// EventCreateResponse là DTO trả về sau khi tạo event thành công.
type EventCreateResponse struct {
	ID            uuid.UUID `json:"id"`
	ProductItemID uuid.UUID `json:"product_item_id"`
	EventType     string    `json:"event_type"`
	OccurredAt    time.Time `json:"occurred_at"`
	Location      string    `json:"location,omitempty"`
	Actor         string    `json:"actor,omitempty"`
	Description   string    `json:"description,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
