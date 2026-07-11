package audit_log

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Action constants represent the allowed audit log action types.
const (
	ActionCreate = "CREATE"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"
)

// AuditLog represents a single audit log entry persisted to the database.
// It captures who did what, to which entity, and what the data looked like
// before and after the operation.
type AuditLog struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID    *uuid.UUID     `gorm:"type:uuid"`
	Action    string         `gorm:"type:varchar(20);not null"`
	Entity    string         `gorm:"type:varchar(50);not null"`
	EntityID  uuid.UUID      `gorm:"type:uuid;not null"`
	OldData   datatypes.JSON `gorm:"type:jsonb"`
	NewData   datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt time.Time      `gorm:"not null;default:now()"`
}

// TableName overrides the default GORM table name.
func (AuditLog) TableName() string {
	return "audit_logs"
}
