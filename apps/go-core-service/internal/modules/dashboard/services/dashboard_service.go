package services

import (
	"context"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/dashboard/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/dashboard/repositories"
)

type DashboardService interface {
	GetStats(ctx context.Context) (*dto.DashboardStats, error)
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
