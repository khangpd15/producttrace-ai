package adapters

import (
	"context"

	"github.com/google/uuid"
)

// EventPort defines how events are recorded into the Product Timeline.
type EventPort interface {
	RecordEvent(ctx context.Context, productItemID uuid.UUID, eventType string, payload interface{}) error
}

// AuditLogPort defines how actions are logged for traceability.
type AuditLogPort interface {
	LogAction(ctx context.Context, action string, userID uuid.UUID, resourceID uuid.UUID, details interface{}) error
}

// NotificationPort handles sending notifications to warranty staff or users.
type NotificationPort interface {
	NotifyStaff(ctx context.Context, claimID uuid.UUID, message string) error
}

// OwnershipPort checks the ownership logic.
type OwnershipPort interface {
	VerifyOwnership(ctx context.Context, userID uuid.UUID, productItemID uuid.UUID) (bool, error)
}

// ProductItemPort retrieves product metadata/status to verify warranty validity.
type ProductItemPort interface {
	CheckWarrantyValidity(ctx context.Context, productItemID uuid.UUID) (bool, error)
}
