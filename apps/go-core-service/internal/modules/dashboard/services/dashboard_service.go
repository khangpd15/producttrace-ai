package services

import (
	"context"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/dashboard/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/dashboard/repositories"
)

type DashboardService interface {
	GetStats(ctx context.Context) (*dto.DashboardStats, error)
	GetActivities(ctx context.Context, limit int) ([]*dto.DashboardActivity, error)
	GetAlerts(ctx context.Context) ([]*dto.DashboardAlert, error)
	GetProductionSalesChart(ctx context.Context) ([]*dto.DashboardChartItem, error)
}

type dashboardService struct {
	dashboardRepo repositories.DashboardRepository
}

func NewDashboardService(dashboardRepo repositories.DashboardRepository) DashboardService {
	return &dashboardService{
		dashboardRepo: dashboardRepo,
	}
}

func (s *dashboardService) GetStats(ctx context.Context) (*dto.DashboardStats, error) {
	return s.dashboardRepo.GetStats(ctx)
}

func (s *dashboardService) GetActivities(ctx context.Context, limit int) ([]*dto.DashboardActivity, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.dashboardRepo.GetActivities(ctx, limit)
}

func (s *dashboardService) GetAlerts(ctx context.Context) ([]*dto.DashboardAlert, error) {
	return s.dashboardRepo.GetAlerts(ctx)
}

func (s *dashboardService) GetProductionSalesChart(ctx context.Context) ([]*dto.DashboardChartItem, error) {
	return s.dashboardRepo.GetProductionSalesChart(ctx)
}
