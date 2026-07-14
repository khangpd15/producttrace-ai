package audit_log

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// AuditLogRepository defines the data-access contract for audit_logs.
// Implementations must only perform CRUD operations — no business logic allowed.
type AuditLogRepository interface {
	// Create inserts a new AuditLog record into the database.
	Create(ctx context.Context, log *AuditLog) error
	
	// Find retrieves a list of AuditLogs based on filters.
	Find(ctx context.Context, action, entity string, fromDate, toDate *time.Time, page, limit int) ([]*AuditLog, int64, error)
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

func (r *auditLogRepository) Find(ctx context.Context, action, entity string, fromDate, toDate *time.Time, page, limit int) ([]*AuditLog, int64, error) {
	query := r.db.WithContext(ctx).Model(&AuditLog{})

	if action != "" {
		query = query.Where("action = ?", action)
	}
	if entity != "" {
		query = query.Where("entity = ?", entity)
	}
	if fromDate != nil {
		query = query.Where("created_at >= ?", fromDate)
	}
	if toDate != nil {
		query = query.Where("created_at <= ?", toDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if page > 0 && limit > 0 {
		query = query.Offset((page - 1) * limit)
	}

	var logs []*AuditLog
	// Order by created_at desc
	if err := query.Order("created_at desc").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
