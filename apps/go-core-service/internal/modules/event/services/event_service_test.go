package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/event/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/event/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/event/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/event/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// Mock EventRepository
// ---------------------------------------------------------------------------

type mockEventRepository struct {
	mock.Mock
}

func (m *mockEventRepository) Create(ctx context.Context, ev *entities.Event) (*response.EventCreateResponse, error) {
	args := m.Called(ctx, ev)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.EventCreateResponse), args.Error(1)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func buildCreateEventRequest(occurredAt *time.Time) *request.CreateEventRequest {
	return &request.CreateEventRequest{
		ProductItemID: uuid.New(),
		EventType:     "SHIPPED",
		OccurredAt:    occurredAt,
		Location:      "Hanoi",
		Actor:         "warehouse-bot",
		Description:   "Item shipped from warehouse",
		Metadata:      map[string]string{"carrier": "GHN"},
	}
}

func buildEventCreateResponse(productItemID uuid.UUID, occurredAt time.Time) *response.EventCreateResponse {
	return &response.EventCreateResponse{
		ID:            uuid.New(),
		ProductItemID: productItemID,
		EventType:     "SHIPPED",
		OccurredAt:    occurredAt,
		Location:      "Hanoi",
		Actor:         "warehouse-bot",
		Description:   "Item shipped from warehouse",
		CreatedAt:     time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestCreateEvent_WithOccurredAt_Success kiểm tra luồng thành công khi
// OccurredAt được cung cấp. Repo phải nhận đúng entity và trả về response.
func TestCreateEvent_WithOccurredAt_Success(t *testing.T) {
	t.Parallel()

	repoMock := new(mockEventRepository)
	svc := services.NewEventService(repoMock)

	fixedTime := time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC)
	req := buildCreateEventRequest(&fixedTime)

	expectedResp := buildEventCreateResponse(req.ProductItemID, fixedTime)

	// Repo sẽ được gọi 1 lần với bất kỳ entity nào (ID được sinh ra bên trong service).
	repoMock.
		On("Create", mock.Anything, mock.MatchedBy(func(ev *entities.Event) bool {
			return ev.ProductItemID == req.ProductItemID &&
				ev.EventType == req.EventType &&
				ev.OccurredAt.Equal(fixedTime) &&
				ev.Location == req.Location &&
				ev.Actor == req.Actor &&
				ev.Description == req.Description
		})).
		Return(expectedResp, nil).
		Once()

	resp, err := svc.CreateEvent(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedResp.ID, resp.ID)
	assert.Equal(t, expectedResp.ProductItemID, resp.ProductItemID)
	assert.Equal(t, expectedResp.EventType, resp.EventType)
	assert.WithinDuration(t, fixedTime, resp.OccurredAt, time.Second)

	repoMock.AssertExpectations(t)
}

// TestCreateEvent_OccurredAtNil_DefaultsToNow kiểm tra rằng khi OccurredAt
// là nil, service tự gán thời điểm hiện tại trước khi gọi mapper/repo.
func TestCreateEvent_OccurredAtNil_DefaultsToNow(t *testing.T) {
	t.Parallel()

	repoMock := new(mockEventRepository)
	svc := services.NewEventService(repoMock)

	req := buildCreateEventRequest(nil) // OccurredAt = nil

	before := time.Now().Add(-time.Second)

	repoMock.
		On("Create", mock.Anything, mock.MatchedBy(func(ev *entities.Event) bool {
			// OccurredAt phải được set và nằm trong khoảng [before, now+1s].
			return !ev.OccurredAt.IsZero() && ev.OccurredAt.After(before)
		})).
		Return(buildEventCreateResponse(req.ProductItemID, time.Now()), nil).
		Once()

	resp, err := svc.CreateEvent(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	// OccurredAt trong request phải được điền (không còn nil sau khi service xử lý).
	assert.NotNil(t, req.OccurredAt)

	repoMock.AssertExpectations(t)
}

// TestCreateEvent_RepoError_ReturnsError kiểm tra rằng khi repo trả về lỗi,
// service trả về nil response và đúng error đó.
func TestCreateEvent_RepoError_ReturnsError(t *testing.T) {
	t.Parallel()

	repoMock := new(mockEventRepository)
	svc := services.NewEventService(repoMock)

	req := buildCreateEventRequest(nil)
	dbErr := errors.New("connection refused")

	repoMock.
		On("Create", mock.Anything, mock.AnythingOfType("*entities.Event")).
		Return(nil, dbErr).
		Once()

	resp, err := svc.CreateEvent(context.Background(), req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, dbErr, err)

	repoMock.AssertExpectations(t)
}

// TestCreateEvent_EntityIDIsUnique kiểm tra rằng mỗi lần gọi CreateEvent
// đều tạo ra một UUID mới khác nhau cho entity (không bị reuse).
func TestCreateEvent_EntityIDIsUnique(t *testing.T) {
	t.Parallel()

	repoMock := new(mockEventRepository)
	svc := services.NewEventService(repoMock)

	capturedIDs := make([]uuid.UUID, 0, 2)

	repoMock.
		On("Create", mock.Anything, mock.MatchedBy(func(ev *entities.Event) bool {
			capturedIDs = append(capturedIDs, ev.ID)
			return true
		})).
		Return(buildEventCreateResponse(uuid.New(), time.Now()), nil)

	req1 := buildCreateEventRequest(nil)
	req2 := buildCreateEventRequest(nil)

	_, _ = svc.CreateEvent(context.Background(), req1)
	_, _ = svc.CreateEvent(context.Background(), req2)

	assert.Len(t, capturedIDs, 2)
	assert.NotEqual(t, capturedIDs[0], capturedIDs[1], "mỗi lần gọi phải sinh UUID khác nhau")
}

// TestCreateEvent_MetadataNil_NoError đảm bảo rằng khi Metadata là nil,
// service vẫn hoạt động bình thường (không panic ở mapper).
func TestCreateEvent_MetadataNil_NoError(t *testing.T) {
	t.Parallel()

	repoMock := new(mockEventRepository)
	svc := services.NewEventService(repoMock)

	req := &request.CreateEventRequest{
		ProductItemID: uuid.New(),
		EventType:     "PRODUCED",
		Metadata:      nil,
	}

	repoMock.
		On("Create", mock.Anything, mock.MatchedBy(func(ev *entities.Event) bool {
			return ev.MetadataJSON == nil
		})).
		Return(buildEventCreateResponse(req.ProductItemID, time.Now()), nil).
		Once()

	resp, err := svc.CreateEvent(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	repoMock.AssertExpectations(t)
}

// TestCreateEvent_ContextPropagated kiểm tra rằng context được truyền thẳng
// xuống repo mà không bị thay thế hay bỏ qua.
func TestCreateEvent_ContextPropagated(t *testing.T) {
	t.Parallel()

	repoMock := new(mockEventRepository)
	svc := services.NewEventService(repoMock)

	type ctxKey string
	ctx := context.WithValue(context.Background(), ctxKey("trace-id"), "abc-123")

	req := buildCreateEventRequest(nil)

	repoMock.
		On("Create", mock.MatchedBy(func(c context.Context) bool {
			return c.Value(ctxKey("trace-id")) == "abc-123"
		}), mock.AnythingOfType("*entities.Event")).
		Return(buildEventCreateResponse(req.ProductItemID, time.Now()), nil).
		Once()

	resp, err := svc.CreateEvent(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	repoMock.AssertExpectations(t)
}
