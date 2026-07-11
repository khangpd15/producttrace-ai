package services

import (
	"context"
	"errors"
	"testing"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/dashboard/dto"
	"github.com/stretchr/testify/assert"
)

type mockDashboardRepository struct {
	getStatsFn func(ctx context.Context) (*dto.DashboardStats, error)
}

func (m *mockDashboardRepository) GetStats(ctx context.Context) (*dto.DashboardStats, error) {
	return m.getStatsFn(ctx)
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
