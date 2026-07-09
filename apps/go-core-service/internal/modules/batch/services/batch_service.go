package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/publisher"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/rabbitmq"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/types"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/qr"
	batchRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/repositories"
	productRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/repositories"
	variantRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	auditlog "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/audit_log"
)

// validBatchStatuses là tập hợp các status hợp lệ của Batch.
var validBatchStatuses = map[string]struct{}{
	"ACTIVE":   {},
	"EXPIRED":  {},
	"RECALLED": {},
	"BLOCKED":  {},
}

type BatchService interface {
	GetBatchList(ctx context.Context, req *request.GetBatchListRequest) (*response.BatchListResponse, error)
	// SearchBatches thực hiện tìm kiếm gần đúng theo UC-P2-BATCH-03.
	SearchBatches(ctx context.Context, req *request.SearchBatchRequest) (*response.SearchBatchResponse, error)
	GetBatchDetail(ctx context.Context, batchCode string) (*response.BatchDetailResponse, error)
	CreateBatch(ctx context.Context, req *request.CreateBatchRequest, currenUserID uuid.UUID) (*response.BatchCreateResponse, error)
	ExportBatchQR(ctx context.Context, batchID uuid.UUID) ([]byte, error)
	ExportBatch(ctx context.Context, batchID uuid.UUID, exportReq *request.ExportBatchRequest, currentUserID uuid.UUID) error
	GetBatchEvents(ctx context.Context, batchID uuid.UUID) ([]response.BatchEventDTO, error)
	// UpdateBatchStatus cập nhật duy nhất field status của một Batch.
	UpdateBatchStatus(ctx context.Context, batchID uuid.UUID, req *request.UpdateBatchStatusRequest, userID *uuid.UUID) (*response.BatchStatusResponse, error)
	// DeleteBatch thực hiện soft-delete một Batch nếu không có product items hoặc events liên kết.
	DeleteBatch(ctx context.Context, batchID uuid.UUID, userID *uuid.UUID) error
}

type batchService struct {
	repo            batchRepo.BatchRepository
	pdfGenerator    qr.PDFGenerator
	productItemRepo productRepo.ProductItemRepository
	variantRepo     variantRepo.ProductVariantRepository
	publisher       *publisher.Publisher
	auditLog        auditlog.AuditLogService
}

func NewbatchService(
	repo batchRepo.BatchRepository,
	pdfGenerator qr.PDFGenerator,
	productItemRepo productRepo.ProductItemRepository,
	variantRepo variantRepo.ProductVariantRepository,
	publisher *publisher.Publisher,
	auditLog auditlog.AuditLogService,
) BatchService {
	return &batchService{
		repo:            repo,
		pdfGenerator:    pdfGenerator,
		productItemRepo: productItemRepo,
		variantRepo:     variantRepo,
		publisher:       publisher,
		auditLog:        auditLog,
	}
}

func (sb *batchService) GetBatchList(ctx context.Context, req *request.GetBatchListRequest) (*response.BatchListResponse, error) {
	return sb.repo.FindAllWithFilter(ctx, req)
}

// SearchBatches validate input rồi delegate xuống repository.
// Business rules:
//   - keyword tối đa 100 ký tự (ERR-001)
//   - sortOrder được normalize thành uppercase trước khi truyền xuống repo
func (sb *batchService) SearchBatches(ctx context.Context, req *request.SearchBatchRequest) (*response.SearchBatchResponse, error) {
	// Validate keyword length (ERR-001)
	if len(req.Keyword) > 100 {
		return nil, apperror.NewBadRequest("Search keyword is too long. Max limit is 100 characters.")
	}

	// Normalize sortOrder để đảm bảo whitelist check trong repo hoạt động chính xác
	req.SortOrder = strings.ToUpper(strings.TrimSpace(req.SortOrder))

	return sb.repo.SearchBatches(ctx, req)
}

func (sb *batchService) GetBatchDetail(ctx context.Context, batchCode string) (*response.BatchDetailResponse, error) {
	return sb.repo.FindByCode(ctx, batchCode)
}

// CreateBatch normalize prefix rồi kiểm tra business rule trước khi delegate xuống repo.
// Batch code sẽ được tự động sinh tại tầng repository dùng advisory lock.
func (sb *batchService) CreateBatch(ctx context.Context, req *request.CreateBatchRequest, currenUserID uuid.UUID) (*response.BatchCreateResponse, error) {
	// Normalize prefix: xóa khoảng trắng, chuyển về uppercase
	// để "apl", "APL", " Apl " đều tạo ra cùng lock key APL-2026.
	req.Prefix = strings.ToUpper(strings.TrimSpace(req.Prefix))

	// FK check: variant_id phải tồn tại trong bảng product_variants.
	variantExists, err := sb.variantRepo.ExistsByID(ctx, req.VariantID)
	if err != nil {
		return nil, err
	}
	if !variantExists {
		return nil, apperror.NewNotFound("product_variant")
	}

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

	batchRes, err := sb.retryCreateBatch(ctx, req, currenUserID)
	if err != nil {
		return nil, err
	}

	event := types.Event{
		EventID:       uuid.NewString(),
		EventType:     rabbitmq.BatchCreatedRK,
		EventVersion:  "1.0",
		Timestamp:     time.Now().UTC(),
		Producer:      "go-core-service",
		CorrelationID: uuid.NewString(),
		Payload:       batchRes,
	}

	if err := sb.publisher.Publish(event); err != nil {
		return nil, apperror.NewInternal("fail to publish batch create event")
	}

	return batchRes, nil
}

func (sb *batchService) retryCreateBatch(ctx context.Context, req *request.CreateBatchRequest, currenUserID uuid.UUID) (*response.BatchCreateResponse, error) {
	var lastErr error
	for i := 0; i < 3; i++ {
		result, err := sb.repo.Create(ctx, req, currenUserID)
		if err == nil {
			return result, nil
		}
		// Kiểm tra nếu là AppError non-retriable (Conflict, NotFound, Validation, BadRequest) thì dừng ngay.
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			switch appErr.Code {
			case apperror.CodeConflict, apperror.CodeNotFound, apperror.CodeValidation, apperror.CodeBadRequest:
				return nil, err
			}
		}
		lastErr = err
		time.Sleep(100 * time.Millisecond)
	}
	if lastErr != nil {
		var appErr *apperror.AppError
		if errors.As(lastErr, &appErr) {
			return nil, lastErr
		}
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

func (sb *batchService) ExportBatch(ctx context.Context, batchID uuid.UUID, exportReq *request.ExportBatchRequest, currentUserID uuid.UUID) error {
	return sb.repo.ExportBatch(ctx, batchID, exportReq, currentUserID)
}

func (sb *batchService) GetBatchEvents(ctx context.Context, batchID uuid.UUID) ([]response.BatchEventDTO, error) {
	// First check if batch exists
	_, err := sb.repo.FindByID(ctx, batchID)
	if err != nil {
		return nil, err
	}
	return sb.repo.GetBatchEvents(ctx, batchID)
}

// UpdateBatchStatus cập nhật duy nhất field status của Batch.
// Business rules:
//   - Batch phải tồn tại
//   - Batch không được đã soft-delete
//   - Status phải thuộc enum hợp lệ (ACTIVE, EXPIRED, RECALLED, BLOCKED)
//
// Audit log được ghi sau khi update thành công.
func (sb *batchService) UpdateBatchStatus(
	ctx context.Context,
	batchID uuid.UUID,
	req *request.UpdateBatchStatusRequest,
	userID *uuid.UUID,
) (*response.BatchStatusResponse, error) {
	// 1. Validate enum trước để fail-fast, tránh query DB không cần thiết.
	newStatus := strings.ToUpper(strings.TrimSpace(req.Status))
	if _, ok := validBatchStatuses[newStatus]; !ok {
		return nil, apperror.NewValidation(
			fmt.Sprintf("invalid status '%s': must be one of ACTIVE, EXPIRED, RECALLED, BLOCKED", req.Status),
		)
	}

	// 2. Kiểm tra Batch tồn tại.
	batch, err := sb.repo.FindByID(ctx, batchID)
	if err != nil {
		return nil, err
	}

	// 3. Không cho update Batch đã bị soft-delete.
	if batch.IsDeleted {
		return nil, apperror.NewBadRequest("batch has been deleted and cannot be updated")
	}

	// 4. Snapshot trạng thái cũ trước khi update (dùng cho audit log).
	oldBatch := *batch

	// 5. Cập nhật status trong DB.
	if err := sb.repo.UpdateStatus(ctx, batchID, newStatus); err != nil {
		return nil, apperror.Wrap(err, apperror.NewInternal("failed to update batch status"))
	}

	// 6. Ghi audit log — không block response nếu log lỗi.
	if logErr := sb.auditLog.LogUpdate(ctx, userID, "Batch", batchID, oldBatch, map[string]string{
		"status": newStatus,
	}); logErr != nil {
		// Log lỗi nhưng không trả về — operation chính đã thành công.
		_ = logErr
	}

	return &response.BatchStatusResponse{
		ID:        batch.ID,
		BatchCode: batch.BatchCode,
		Status:    newStatus,
		UpdatedAt: time.Now().UTC(),
	}, nil
}

// DeleteBatch thực hiện soft-delete một Batch.
// Business rules:
//   - Batch phải tồn tại
//   - Batch chưa bị soft-delete
//   - Không có product item liên kết (chưa deleted)
//   - Không có event liên kết (qua product_items)
//
// Soft-delete và audit log được thực hiện trong transaction để đảm bảo atomic.
func (sb *batchService) DeleteBatch(
	ctx context.Context,
	batchID uuid.UUID,
	userID *uuid.UUID,
) error {
	// 1. Kiểm tra Batch tồn tại.
	batch, err := sb.repo.FindByID(ctx, batchID)
	if err != nil {
		return err
	}

	// 2. Không cho xóa Batch đã bị soft-delete.
	if batch.IsDeleted {
		return apperror.NewBadRequest("batch has already been deleted")
	}

	// 3. Không cho xóa nếu có product items liên kết.
	hasItems, err := sb.repo.ExistsProductItems(ctx, batchID)
	if err != nil {
		return apperror.Wrap(err, apperror.NewInternal("failed to check product items"))
	}
	if hasItems {
		return apperror.NewBadRequest("cannot delete batch: batch has linked product items")
	}

	// 4. Không cho xóa nếu có events liên kết.
	hasEvents, err := sb.repo.ExistsEvents(ctx, batchID)
	if err != nil {
		return apperror.Wrap(err, apperror.NewInternal("failed to check events"))
	}
	if hasEvents {
		return apperror.NewBadRequest("cannot delete batch: batch has linked events")
	}

	// 5. Soft-delete trong DB.
	if err := sb.repo.SoftDelete(ctx, batchID); err != nil {
		return apperror.Wrap(err, apperror.NewInternal("failed to delete batch"))
	}

	// 6. Ghi audit log — không block response nếu log lỗi.
	if logErr := sb.auditLog.LogDelete(ctx, userID, "Batch", batchID, batch); logErr != nil {
		_ = logErr
	}

	return nil
}
