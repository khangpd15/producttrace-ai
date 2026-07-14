package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/dashboard/dto"
	"github.com/stretchr/testify/assert"
)

type mockDashboardRepository struct {
	getStatsFn                func(ctx context.Context) (*dto.DashboardStats, error)
	getActivitiesFn           func(ctx context.Context, limit int) ([]*dto.DashboardActivity, error)
	getAlertsFn               func(ctx context.Context) ([]*dto.DashboardAlert, error)
	getProductionSalesChartFn func(ctx context.Context) ([]*dto.DashboardChartItem, error)
}

func (m *mockDashboardRepository) GetStats(ctx context.Context) (*dto.DashboardStats, error) {
	return m.getStatsFn(ctx)
}

func (m *mockDashboardRepository) GetActivities(ctx context.Context, limit int) ([]*dto.DashboardActivity, error) {
	return m.getActivitiesFn(ctx, limit)
}

func (m *mockDashboardRepository) GetAlerts(ctx context.Context) ([]*dto.DashboardAlert, error) {
	return m.getAlertsFn(ctx)
}

func (m *mockDashboardRepository) GetProductionSalesChart(ctx context.Context) ([]*dto.DashboardChartItem, error) {
	return m.getProductionSalesChartFn(ctx)
}

func TestDashboardService_GetStats_Success(t *testing.T) {
	expectedStats := &dto.DashboardStats{
		TotalProducts:        100,
		TotalBatches:         50,
		TotalOwnerships:      200,
		TotalUnderWarranty:   80,
		TotalPendingApproval: 10,
		TotalLocations:       5,
	}

	mockRepo := &mockDashboardRepository{
		getStatsFn: func(ctx context.Context) (*dto.DashboardStats, error) {
			return expectedStats, nil
		},
	}

	service := NewDashboardService(mockRepo)
	stats, err := service.GetStats(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedStats, stats)
}

func TestDashboardService_GetStats_Error(t *testing.T) {
	expectedErr := errors.New("database query error")

	mockRepo := &mockDashboardRepository{
		getStatsFn: func(ctx context.Context) (*dto.DashboardStats, error) {
			return nil, expectedErr
		},
	}

	service := NewDashboardService(mockRepo)
	stats, err := service.GetStats(context.Background())

	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Equal(t, expectedErr, err)
}

func TestDashboardService_GetActivities_Success(t *testing.T) {
	now := time.Now()
	expectedActivities := []*dto.DashboardActivity{
		{
			ID:          "1",
			EventType:   "WAREHOUSE_IN",
			Title:       "Nhập kho",
			Description: "Lô hàng BATCH-2026-KG01 được nhập kho thành công.",
			CreatedAt:   now,
		},
	}

	mockRepo := &mockDashboardRepository{
		getActivitiesFn: func(ctx context.Context, limit int) ([]*dto.DashboardActivity, error) {
			assert.Equal(t, 5, limit)
			return expectedActivities, nil
		},
	}

	service := NewDashboardService(mockRepo)
	activities, err := service.GetActivities(context.Background(), 5)

	assert.NoError(t, err)
	assert.Equal(t, expectedActivities, activities)
}

func TestDashboardService_GetActivities_DefaultLimit(t *testing.T) {
	expectedActivities := []*dto.DashboardActivity{}

	mockRepo := &mockDashboardRepository{
		getActivitiesFn: func(ctx context.Context, limit int) ([]*dto.DashboardActivity, error) {
			assert.Equal(t, 10, limit)
			return expectedActivities, nil
		},
	}

	service := NewDashboardService(mockRepo)
	activities, err := service.GetActivities(context.Background(), 0)

	assert.NoError(t, err)
	assert.Equal(t, expectedActivities, activities)
}

func TestDashboardService_GetActivities_Error(t *testing.T) {
	expectedErr := errors.New("query error")

	mockRepo := &mockDashboardRepository{
		getActivitiesFn: func(ctx context.Context, limit int) ([]*dto.DashboardActivity, error) {
			return nil, expectedErr
		},
	}

	service := NewDashboardService(mockRepo)
	activities, err := service.GetActivities(context.Background(), 10)

	assert.Error(t, err)
	assert.Nil(t, activities)
	assert.Equal(t, expectedErr, err)
}

func TestDashboardService_GetAlerts_Success(t *testing.T) {
	expectedAlerts := []*dto.DashboardAlert{
		{
			ID:          "batch-123",
			Type:        "DANGER",
			Title:       "Phát hiện lô sản phẩm hết hạn",
			Description: "Lô hàng BATCH-2026-SP12 (Sơn Spec) đã hết hạn sử dụng. Yêu cầu khóa mã QR.",
		},
	}

	mockRepo := &mockDashboardRepository{
		getAlertsFn: func(ctx context.Context) ([]*dto.DashboardAlert, error) {
			return expectedAlerts, nil
		},
	}

	service := NewDashboardService(mockRepo)
	alerts, err := service.GetAlerts(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedAlerts, alerts)
}

func TestDashboardService_GetAlerts_Error(t *testing.T) {
	expectedErr := errors.New("query error")

	mockRepo := &mockDashboardRepository{
		getAlertsFn: func(ctx context.Context) ([]*dto.DashboardAlert, error) {
			return nil, expectedErr
		},
	}

	service := NewDashboardService(mockRepo)
	alerts, err := service.GetAlerts(context.Background())

	assert.Error(t, err)
	assert.Nil(t, alerts)
	assert.Equal(t, expectedErr, err)
}

func TestDashboardService_GetProductionSalesChart_Success(t *testing.T) {
	expectedChart := []*dto.DashboardChartItem{
		{
			TimePeriod:       time.Now(),
			ProductionVolume: 45.2,
			SalesVolume:      38.1,
		},
	}

	mockRepo := &mockDashboardRepository{
		getProductionSalesChartFn: func(ctx context.Context) ([]*dto.DashboardChartItem, error) {
			return expectedChart, nil
		},
	}

	service := NewDashboardService(mockRepo)
	chart, err := service.GetProductionSalesChart(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedChart, chart)
}

func TestDashboardService_GetProductionSalesChart_Error(t *testing.T) {
	expectedErr := errors.New("chart query error")

	mockRepo := &mockDashboardRepository{
		getProductionSalesChartFn: func(ctx context.Context) ([]*dto.DashboardChartItem, error) {
			return nil, expectedErr
		},
	}

	service := NewDashboardService(mockRepo)
	chart, err := service.GetProductionSalesChart(context.Background())

	assert.Error(t, err)
	assert.Nil(t, chart)
	assert.Equal(t, expectedErr, err)
}
