package services

import (
	"context"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/repositories"
)

type BatchService interface {
	GetBatchList(ctx context.Context) ([]*response.BatchListResponse, error)
	GetBatchDetail(ctx context.Context, batchCode string) (*response.BatchDetailResponse, error)
}

type batchService struct {
	repo repositories.BatchRepository
}

func NewbatchService(repo repositories.BatchRepository) BatchService {
	return &batchService{
		repo: repo,
	}
}

func (sb *batchService) GetBatchList(ctx context.Context) ([]*response.BatchListResponse, error) {
	return sb.repo.FindAll(ctx)
}

func (sb *batchService) GetBatchDetail(ctx context.Context, batchCode string) (*response.BatchDetailResponse, error) {
	return sb.repo.FindByCode(ctx, batchCode)
}
