package mapper

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/event/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/event/entities"
	"gorm.io/datatypes"
)

// CreateEventRequestToEntity chuyển đổi CreateEventRequest thành Event entity
// sẵn sàng để insert vào database.
func CreateEventRequestToEntity(newID uuid.UUID, req *request.CreateEventRequest) *entities.Event {
	occurredAt := time.Now()
	if req.OccurredAt != nil {
		occurredAt = *req.OccurredAt
	}

	var metaJSON datatypes.JSON
	if req.Metadata != nil {
		if b, err := json.Marshal(req.Metadata); err == nil {
			metaJSON = datatypes.JSON(b)
		}
	}

	return &entities.Event{
		ID:            newID,
		ProductItemID: req.ProductItemID,
		EventType:     req.EventType,
		OccurredAt:    occurredAt,
		Location:      req.Location,
		Actor:         req.Actor,
		Description:   req.Description,
		MetadataJSON:  metaJSON,
	}
}
