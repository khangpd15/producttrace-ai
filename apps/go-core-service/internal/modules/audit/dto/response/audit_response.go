package response

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AuditLogResponse struct {
	ID        uuid.UUID      `json:"id"`
	UserID    *uuid.UUID     `json:"user_id"`
	Action    string         `json:"action"`
	Entity    string         `json:"entity"`
	EntityID  uuid.UUID      `json:"entity_id"`
	OldData   datatypes.JSON `json:"old_data,omitempty"`
	NewData   datatypes.JSON `json:"new_data,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}
