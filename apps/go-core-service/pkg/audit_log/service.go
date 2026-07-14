package audit_log

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// AuditLogService defines the business-logic contract for recording audit events.
// Any module (ProductService, BatchService, etc.) can depend on this interface
// and call the appropriate helper methods after each mutating operation.
type AuditLogService interface {
	// Log is the primary method that marshals oldData/newData, builds an AuditLog
	// model, and delegates persistence to the repository.
	// Pass nil for oldData (CREATE) or newData (DELETE) to store SQL NULL.
	Log(
		ctx context.Context,
		userID *uuid.UUID,
		action string,
		entity string,
		entityID uuid.UUID,
		oldData any,
		newData any,
	) error

	// --- Convenience helpers (wrap Log) ---

	// LogCreate records a CREATE action.
	// newData is the newly created entity; oldData is always NULL.
	LogCreate(ctx context.Context, userID *uuid.UUID, entity string, entityID uuid.UUID, newData any) error

	// LogUpdate records an UPDATE action.
	// oldData is a snapshot taken before the update; newData is the updated entity.
	LogUpdate(ctx context.Context, userID *uuid.UUID, entity string, entityID uuid.UUID, oldData any, newData any) error

	// LogDelete records a DELETE action.
	// oldData is the entity that was deleted; newData is always NULL.
	LogDelete(ctx context.Context, userID *uuid.UUID, entity string, entityID uuid.UUID, oldData any) error

	// GetLogs retrieves audit logs based on filters.
	GetLogs(ctx context.Context, action, entity string, fromDate, toDate *time.Time, page, limit int) ([]*AuditLog, int64, error)
}

type auditLogService struct {
	repo AuditLogRepository
}

// NewAuditLogService creates a new AuditLogService with the given repository.
// Inject this into any module that needs to record audit history.
func NewAuditLogService(repo AuditLogRepository) AuditLogService {
	return &auditLogService{repo: repo}
}

// Log marshals oldData and newData to JSON, constructs an AuditLog model,
// generates a UUID, sets CreatedAt, and delegates persistence to the repository.
//
// If oldData is nil  → old_data column is stored as SQL NULL.
// If newData is nil  → new_data column is stored as SQL NULL.
// If marshaling fails → the error is returned immediately; nothing is persisted.
func (s *auditLogService) Log(
	ctx context.Context,
	userID *uuid.UUID,
	action string,
	entity string,
	entityID uuid.UUID,
	oldData any,
	newData any,
) error {
	oldJSON, err := marshalJSON(oldData)
	if err != nil {
		return fmt.Errorf("audit_log: failed to marshal oldData for entity %s/%s: %w", entity, entityID, err)
	}

	newJSON, err := marshalJSON(newData)
	if err != nil {
		return fmt.Errorf("audit_log: failed to marshal newData for entity %s/%s: %w", entity, entityID, err)
	}

	entry := &AuditLog{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		OldData:   oldJSON,
		NewData:   newJSON,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, entry); err != nil {
		return fmt.Errorf("audit_log: failed to persist log for entity %s/%s: %w", entity, entityID, err)
	}

	return nil
}

// LogCreate is a convenience wrapper that records a CREATE action.
// oldData is always nil (NULL) for create operations.
func (s *auditLogService) LogCreate(
	ctx context.Context,
	userID *uuid.UUID,
	entity string,
	entityID uuid.UUID,
	newData any,
) error {
	return s.Log(ctx, userID, ActionCreate, entity, entityID, nil, newData)
}

// LogUpdate is a convenience wrapper that records an UPDATE action.
// Caller must snapshot the entity state before mutating it and pass it as oldData.
func (s *auditLogService) LogUpdate(
	ctx context.Context,
	userID *uuid.UUID,
	entity string,
	entityID uuid.UUID,
	oldData any,
	newData any,
) error {
	return s.Log(ctx, userID, ActionUpdate, entity, entityID, oldData, newData)
}

// LogDelete is a convenience wrapper that records a DELETE action.
// newData is always nil (NULL) for delete operations.
func (s *auditLogService) LogDelete(
	ctx context.Context,
	userID *uuid.UUID,
	entity string,
	entityID uuid.UUID,
	oldData any,
) error {
	return s.Log(ctx, userID, ActionDelete, entity, entityID, oldData, nil)
}

// marshalJSON converts v to datatypes.JSON.
// Returns nil (SQL NULL) when v is nil, or an error if marshaling fails.
func marshalJSON(v any) (datatypes.JSON, error) {
	if v == nil {
		return nil, nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return datatypes.JSON(b), nil
}

func (s *auditLogService) GetLogs(ctx context.Context, action, entity string, fromDate, toDate *time.Time, page, limit int) ([]*AuditLog, int64, error) {
	return s.repo.Find(ctx, action, entity, fromDate, toDate, page, limit)
}
