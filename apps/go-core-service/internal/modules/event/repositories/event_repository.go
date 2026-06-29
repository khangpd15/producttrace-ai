package repositories

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/event/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/event/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"gorm.io/gorm"
)

type EventRepository interface {
	Create(ctx context.Context, ev *entities.Event) (*response.EventCreateResponse, error)
}

type eventRepository struct {
	db    *gorm.DB
	sqlDB *sql.DB
}

func NewEventRepository(db *gorm.DB, sqlDB *sql.DB) EventRepository {
	return &eventRepository{
		db:    db,
		sqlDB: sqlDB,
	}
}

// Create tạo một event mới bằng raw SQL với RETURNING để lấy ngay dữ liệu đã lưu.
// Dùng cùng pattern với product_item để tránh N+1 query.
func (r *eventRepository) Create(ctx context.Context, ev *entities.Event) (*response.EventCreateResponse, error) {
	insertSQL := `
		INSERT INTO events (
			id, product_item_id,
			event_type, occurred_at,
			location, actor, description, metadata,
			created_at, updated_at, is_deleted
		) VALUES (
			$1, $2,
			$3, $4,
			$5, $6, $7, $8,
			$9, $9, FALSE
		)
		RETURNING id, product_item_id, event_type, occurred_at, location, actor, description, created_at`

	row := r.sqlDB.QueryRowContext(ctx, insertSQL,
		ev.ID,
		ev.ProductItemID,
		ev.EventType,
		ev.OccurredAt,
		ev.Location,
		ev.Actor,
		ev.Description,
		ev.MetadataJSON,
		time.Now(),
	)

	var result response.EventCreateResponse
	err := row.Scan(
		&result.ID,
		&result.ProductItemID,
		&result.EventType,
		&result.OccurredAt,
		&result.Location,
		&result.Actor,
		&result.Description,
		&result.CreatedAt,
	)
	if err != nil {
		log.Println("[event] insert error:", err)
		return nil, apperror.WrapDBError(err, "events")
	}

	return &result, nil
}
