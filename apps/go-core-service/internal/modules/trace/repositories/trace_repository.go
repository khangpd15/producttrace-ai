package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"gorm.io/gorm"
)

type ProductItemDetail struct {
	ItemID       uuid.UUID  `gorm:"column:item_id"`
	ItemCode     string     `gorm:"column:item_code"`
	SerialNumber string     `gorm:"column:serial_number"`
	Status       string     `gorm:"column:status"`
	ProductName  string     `gorm:"column:product_name"`
	ThumbnailURL string     `gorm:"column:thumbnail_url"`
	BatchID      *uuid.UUID `gorm:"column:batch_id"`
}

type TimelineEvent struct {
	EventID     uuid.UUID `gorm:"column:id"`
	EventType   string    `gorm:"column:event_type"`
	Title       string    `gorm:"column:title"`
	Description string    `gorm:"column:description"`
	Location    string    `gorm:"column:location_name"`
	ActorName   string    `gorm:"column:actor_name"`
	Timestamp   time.Time `gorm:"column:created_at"`
}

type AuditLogDetail struct {
	ID        uuid.UUID `gorm:"column:id"`
	Action    string    `gorm:"column:action"`
	Entity    string    `gorm:"column:entity"`
	EntityID  uuid.UUID `gorm:"column:entity_id"`
	OldData   string    `gorm:"column:old_data"`
	NewData   string    `gorm:"column:new_data"`
	UserEmail string    `gorm:"column:user_email"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

type TraceRepository interface {
	FindProductItemByCode(ctx context.Context, code string) (*ProductItemDetail, error)
	FindProductItemsByBatchID(ctx context.Context, batchID uuid.UUID) ([]*ProductItemDetail, error)
	FindEvents(ctx context.Context, itemID uuid.UUID, batchID *uuid.UUID, fromDate *time.Time, toDate *time.Time, eventTypes []string) ([]TimelineEvent, error)
	FindAuditLogs(ctx context.Context, itemID uuid.UUID, batchID *uuid.UUID) ([]AuditLogDetail, error)
}

type traceRepository struct {
	db *gorm.DB
}

func NewTraceRepository(db *gorm.DB) TraceRepository {
	return &traceRepository{db: db}
}

func (r *traceRepository) FindProductItemByCode(ctx context.Context, code string) (*ProductItemDetail, error) {
	var detail ProductItemDetail
	err := r.db.WithContext(ctx).
		Table("product_items pi").
		Select(`
			pi.id as item_id,
			pi.item_code,
			pi.serial_number,
			pi.status,
			p.name as product_name,
			COALESCE(p.thumbnail_url, '') as thumbnail_url,
			pi.batch_id
		`).
		Joins("JOIN product_variants pv ON pi.variant_id = pv.id").
		Joins("JOIN products p ON pv.product_id = p.id").
		Where("(pi.item_code = ? OR pi.serial_number = ?) AND pi.is_deleted = false", code, code).
		Take(&detail).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, apperror.WrapDBError(err, "product_items")
	}

	return &detail, nil
}

func (r *traceRepository) FindEvents(ctx context.Context, itemID uuid.UUID, batchID *uuid.UUID, fromDate *time.Time, toDate *time.Time, eventTypes []string) ([]TimelineEvent, error) {
	var events []TimelineEvent

	query := r.db.WithContext(ctx).
		Table("events e").
		Select(`
			e.id,
			e.event_type,
			COALESCE(e.title, '') as title,
			COALESCE(e.description, '') as description,
			COALESCE(l.name, '') as location_name,
			COALESCE(u.email, '') as actor_name,
			e.created_at
		`).
		Joins("LEFT JOIN locations l ON e.location_id = l.id").
		Joins("LEFT JOIN users u ON e.actor_id = u.id")

	if batchID != nil {
		query = query.Where("e.product_item_id = ? OR (e.batch_id = ? AND e.product_item_id IS NULL)", itemID, *batchID)
	} else {
		query = query.Where("e.product_item_id = ?", itemID)
	}

	if fromDate != nil {
		query = query.Where("e.created_at >= ?", *fromDate)
	}
	if toDate != nil {
		query = query.Where("e.created_at <= ?", *toDate)
	}
	if len(eventTypes) > 0 {
		query = query.Where("e.event_type IN ?", eventTypes)
	}

	err := query.Order("e.created_at ASC").Scan(&events).Error
	if err != nil {
		return nil, apperror.WrapDBError(err, "events")
	}

	return events, nil
}

func (r *traceRepository) FindAuditLogs(ctx context.Context, itemID uuid.UUID, batchID *uuid.UUID) ([]AuditLogDetail, error) {
	var logs []AuditLogDetail

	query := r.db.WithContext(ctx).
		Table("audit_logs a").
		Select(`
			a.id,
			a.action,
			a.entity,
			a.entity_id,
			COALESCE(a.old_data::text, '') as old_data,
			COALESCE(a.new_data::text, '') as new_data,
			COALESCE(u.email, 'system') as user_email,
			a.created_at
		`).
		Joins("LEFT JOIN users u ON a.user_id = u.id")

	if batchID != nil {
		query = query.Where("(a.entity = 'ProductItem' AND a.entity_id = ?) OR (a.entity = 'Batch' AND a.entity_id = ?)", itemID, *batchID)
	} else {
		query = query.Where("a.entity = 'ProductItem' AND a.entity_id = ?", itemID)
	}

	err := query.Order("a.created_at ASC").Scan(&logs).Error
	if err != nil {
		return nil, apperror.WrapDBError(err, "audit_logs")
	}

	return logs, nil
}

func (r *traceRepository) FindProductItemsByBatchID(ctx context.Context, batchID uuid.UUID) ([]*ProductItemDetail, error) {
	var items []*ProductItemDetail
	err := r.db.WithContext(ctx).
		Table("product_items pi").
		Select(`
			pi.id as item_id,
			pi.item_code,
			pi.serial_number,
			pi.status,
			p.name as product_name,
			COALESCE(p.thumbnail_url, '') as thumbnail_url,
			pi.batch_id
		`).
		Joins("JOIN product_variants pv ON pi.variant_id = pv.id").
		Joins("JOIN products p ON pv.product_id = p.id").
		Where("pi.batch_id = ? AND pi.is_deleted = false", batchID).
		Scan(&items).Error

	if err != nil {
		return nil, apperror.WrapDBError(err, "product_items")
	}

	return items, nil
}
