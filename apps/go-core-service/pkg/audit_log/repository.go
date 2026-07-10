package audit_log

import (
	"context"

	"gorm.io/gorm"
)

// AuditLogRepository defines the data-access contract for audit_logs.
// Implementations must only perform CRUD operations — no business logic allowed.
type AuditLogRepository interface {
	// Create inserts a new AuditLog record into the database.
	Create(ctx context.Context, log *AuditLog) error
}

type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new AuditLogRepository backed by GORM.
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

// Create inserts a single AuditLog row into the database.
// This method performs no business logic — it is a pure INSERT operation.
func (r *auditLogRepository) Create(ctx context.Context, log *AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
