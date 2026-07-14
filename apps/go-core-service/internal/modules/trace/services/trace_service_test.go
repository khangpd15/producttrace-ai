package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/trace/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/trace/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/trace/services"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/audit_log"
)

type mockTraceRepository struct {
	mock.Mock
}

func (m *mockTraceRepository) FindProductItemByCode(ctx context.Context, code string) (*repositories.ProductItemDetail, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.ProductItemDetail), args.Error(1)
}

func (m *mockTraceRepository) FindProductItemsByBatchID(ctx context.Context, batchID uuid.UUID) ([]*repositories.ProductItemDetail, error) {
	args := m.Called(ctx, batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*repositories.ProductItemDetail), args.Error(1)
}

func (m *mockTraceRepository) FindEvents(ctx context.Context, itemID uuid.UUID, batchID *uuid.UUID, fromDate *time.Time, toDate *time.Time, eventTypes []string) ([]repositories.TimelineEvent, error) {
	args := m.Called(ctx, itemID, batchID, fromDate, toDate, eventTypes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.TimelineEvent), args.Error(1)
}

func (m *mockTraceRepository) FindAuditLogs(ctx context.Context, itemID uuid.UUID, batchID *uuid.UUID) ([]repositories.AuditLogDetail, error) {
	args := m.Called(ctx, itemID, batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.AuditLogDetail), args.Error(1)
}

type mockAuditLogService struct {
	mock.Mock
}

func (m *mockAuditLogService) Log(ctx context.Context, userID *uuid.UUID, action string, entity string, entityID uuid.UUID, oldData any, newData any) error {
	args := m.Called(ctx, userID, action, entity, entityID, oldData, newData)
	return args.Error(0)
}
func (m *mockAuditLogService) LogCreate(ctx context.Context, userID *uuid.UUID, entity string, entityID uuid.UUID, newData any) error {
	return nil
}
func (m *mockAuditLogService) LogUpdate(ctx context.Context, userID *uuid.UUID, entity string, entityID uuid.UUID, oldData any, newData any) error {
	return nil
}
func (m *mockAuditLogService) LogDelete(ctx context.Context, userID *uuid.UUID, entity string, entityID uuid.UUID, oldData any) error {
	return nil
}
func (m *mockAuditLogService) GetLogs(ctx context.Context, action, entity string, fromDate, toDate *time.Time, page, limit int) ([]*audit_log.AuditLog, int64, error) {
	args := m.Called(ctx, action, entity, fromDate, toDate, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*audit_log.AuditLog), args.Get(1).(int64), args.Error(2)
}

func TestSearchTimeline_TrimAndUppercase(t *testing.T) {
	t.Parallel()

	repo := new(mockTraceRepository)
	audit := new(mockAuditLogService)
	svc := services.NewTraceService(repo, nil, nil, audit, "http://localhost:8080/")

	itemCode := "PT-MILK-SN0001"
	itemID := uuid.New()
	itemDetail := &repositories.ProductItemDetail{
		ItemID:       itemID,
		ItemCode:     itemCode,
		SerialNumber: "SN-0001",
		Status:       "AVAILABLE",
		ProductName:  "Milk",
	}

	repo.On("FindProductItemByCode", mock.Anything, "PT-MILK-SN0001").Return(itemDetail, nil).Once()
	repo.On("FindEvents", mock.Anything, itemID, (*uuid.UUID)(nil), (*time.Time)(nil), (*time.Time)(nil), mock.Anything).Return([]repositories.TimelineEvent{}, nil).Once()
	audit.On("Log", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	req := &request.TraceSearchRequest{
		Code: "  pt-milk-sn0001  ",
	}
	resp, err := svc.SearchTimeline(context.Background(), req, "CUSTOMER", "127.0.0.1", "test-agent", nil)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	repo.AssertExpectations(t)
}

func TestSearchTimeline_InvalidDateRange(t *testing.T) {
	t.Parallel()

	repo := new(mockTraceRepository)
	audit := new(mockAuditLogService)
	svc := services.NewTraceService(repo, nil, nil, audit, "http://localhost:8080/")

	req := &request.TraceSearchRequest{
		Code:     "PT-MILK-SN0001",
		FromDate: "2026-07-08T00:00:00Z",
		ToDate:   "2026-07-01T00:00:00Z",
	}
	_, err := svc.SearchTimeline(context.Background(), req, "CUSTOMER", "127.0.0.1", "test-agent", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Start date cannot be after end date")
}

func TestSearchTimeline_InvalidEventTypes(t *testing.T) {
	t.Parallel()

	repo := new(mockTraceRepository)
	audit := new(mockAuditLogService)
	svc := services.NewTraceService(repo, nil, nil, audit, "http://localhost:8080/")

	req := &request.TraceSearchRequest{
		Code:       "PT-MILK-SN0001",
		EventTypes: "PRODUCED,INVALID_TYPE",
	}
	_, err := svc.SearchTimeline(context.Background(), req, "CUSTOMER", "127.0.0.1", "test-agent", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid event type filter")
}
