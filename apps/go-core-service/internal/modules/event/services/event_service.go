package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/event/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/event/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/event/mapper"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/event/repositories"
)

type EventService interface {
	CreateEvent(ctx context.Context, req *request.CreateEventRequest) (*response.EventCreateResponse, error)
}

type eventService struct {
	repo repositories.EventRepository
}

func NewEventService(repo repositories.EventRepository) EventService {
	return &eventService{
		repo: repo,
	}
}

// CreateEvent tạo một event mới gắn với product item.
// Nếu OccurredAt không được cung cấp, mặc định dùng thời điểm hiện tại.
func (s *eventService) CreateEvent(ctx context.Context, req *request.CreateEventRequest) (*response.EventCreateResponse, error) {
	if req.OccurredAt == nil {
		now := time.Now()
		req.OccurredAt = &now
	}

	newID := uuid.New()
	ev := mapper.CreateEventRequestToEntity(newID, req)

	return s.repo.Create(ctx, ev)
}
