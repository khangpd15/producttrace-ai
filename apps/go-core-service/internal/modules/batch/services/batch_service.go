package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/qr"
	batchRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/repositories"
	productRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
)

type BatchService interface {
	GetBatchList(ctx context.Context) ([]*response.BatchListResponse, error)
	GetBatchDetail(ctx context.Context, batchCode string) (*response.BatchDetailResponse, error)
	CreateBatch(ctx context.Context, req *request.CreateBatchRequest) (*response.BatchCreateResponse, error)
	ExportBatchQR(ctx context.Context, batchID uuid.UUID) ([]byte, error)
}

type batchService struct {
	repo            batchRepo.BatchRepository
	pdfGenerator    qr.PDFGenerator
	productItemRepo productRepo.ProductItemRepository
}

func NewbatchService(repo batchRepo.BatchRepository, pdfGenerator qr.PDFGenerator, productItemRepo productRepo.ProductItemRepository) BatchService {
	return &batchService{
		repo:            repo,
		pdfGenerator:    pdfGenerator,
		productItemRepo: productItemRepo,
	}
}

func (sb *batchService) GetBatchList(ctx context.Context) ([]*response.BatchListResponse, error) {
	return sb.repo.FindAll(ctx)
}

func (sb *batchService) GetBatchDetail(ctx context.Context, batchCode string) (*response.BatchDetailResponse, error) {
	return sb.repo.FindByCode(ctx, batchCode)
}

// CreateBatch normalize prefix rồi kiểm tra business rule trước khi delegate xuống repo.
// Batch code sẽ được tự động sinh tại tầng repository dùng advisory lock.
func (sb *batchService) CreateBatch(ctx context.Context, req *request.CreateBatchRequest) (*response.BatchCreateResponse, error) {
	// Normalize prefix: xóa khoảng trắng, chuyển về uppercase
	// để "apl", "APL", " Apl " đều tạo ra cùng lock key APL-2026.
	req.Prefix = strings.ToUpper(strings.TrimSpace(req.Prefix))

	// Business rule: expiry_date phải >= manufacture_date.
	// Kiểm tra sớm ở tầng service để trả lỗi rõ ràng hơn là đợi DB reject.
	if req.ExpiryDate != nil && req.ManufactureDate != nil {

		if req.ManufactureDate.After(time.Now()) {
			return nil, apperror.NewValidation("manufacture_date must not be in the future")
		}

		if req.ExpiryDate.Before(*req.ManufactureDate) {
			return nil, apperror.NewValidation("expiry_date must be greater than or equal to manufacture_date")
		}
	}

	if req.Quantity <= 0 {
		return nil, apperror.NewValidation("Quantity must be greater than 0")
	}

	return sb.retryCreateBatch(ctx, req)
}

func (sb *batchService) retryCreateBatch(ctx context.Context, req *request.CreateBatchRequest) (*response.BatchCreateResponse, error) {
	for i := 0; i < 3; i++ {
		result, err := sb.repo.Create(ctx, req)
		if err == nil {
			return result, nil
		}
		time.Sleep(100 * time.Millisecond)
		continue
	}
	return nil, apperror.NewInternal("failed to create batch after 3 retries")
}

func (sb *batchService) ExportBatchQR(ctx context.Context, batchID uuid.UUID) ([]byte, error) {

	batch, err := sb.repo.FindByBatchID(ctx, batchID)

	if err != nil {
		return nil, err
	}

	items, err := sb.productItemRepo.
		FindByBatchID(
			ctx,
			batchID,
		)

	if err != nil {
		return nil, err
	}

	labels := make(
		[]qr.ProductItemLabel,
		0,
		len(items),
	)

	for _, item := range items {

		labels = append(
			labels,
			qr.ProductItemLabel{
				ItemCode: item.ItemCode,
				Token:    item.VerificationToken,
			},
		)
	}

	return sb.pdfGenerator.GenerateLabels(
		qr.BatchPDFInput{
			BatchCode: batch.BatchCode,
			Items:     labels,
		},
	)
}
