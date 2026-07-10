package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/response"
	batchEntities "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/qr"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/services"
	itemEntities "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/entities"
	itemReq "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/dto/request"
	itemResp "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/dto/response"
	variantEntities "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockBatchRepository struct {
	mock.Mock
}

func (m *mockBatchRepository) FindAllWithFilter(ctx context.Context, req *request.GetBatchListRequest) (*response.BatchListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.BatchListResponse), args.Error(1)
}

func (m *mockBatchRepository) SearchBatches(ctx context.Context, req *request.SearchBatchRequest) (*response.SearchBatchResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.SearchBatchResponse), args.Error(1)
}

func (m *mockBatchRepository) FindByCode(ctx context.Context, batchCode string) (*response.BatchDetailResponse, error) {
	args := m.Called(ctx, batchCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.BatchDetailResponse), args.Error(1)
}

func (m *mockBatchRepository) FindByBatchID(ctx context.Context, batchID uuid.UUID) (*response.BatchDetailResponse, error) {
	args := m.Called(ctx, batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.BatchDetailResponse), args.Error(1)
}

func (m *mockBatchRepository) GetBatchEvents(ctx context.Context, batchID uuid.UUID) ([]response.BatchEventDTO, error) {
	args := m.Called(ctx, batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.BatchEventDTO), args.Error(1)
}

func (m *mockBatchRepository) FindByID(ctx context.Context, id uuid.UUID) (*batchEntities.Batch, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*batchEntities.Batch), args.Error(1)
}

func (m *mockBatchRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *mockBatchRepository) Create(ctx context.Context, req *request.CreateBatchRequest, currentUserID uuid.UUID) (*response.BatchCreateResponse, error) {
	args := m.Called(ctx, req, currentUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.BatchCreateResponse), args.Error(1)
}

func (m *mockBatchRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *mockBatchRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockBatchRepository) ExportBatch(ctx context.Context, batchID uuid.UUID, exportReq *request.ExportBatchRequest, currentUserID uuid.UUID) error {
	args := m.Called(ctx, batchID, exportReq, currentUserID)
	return args.Error(0)
}

func (m *mockBatchRepository) ExistsProductItems(ctx context.Context, batchID uuid.UUID) (bool, error) {
	args := m.Called(ctx, batchID)
	return args.Bool(0), args.Error(1)
}

func (m *mockBatchRepository) ExistsEvents(ctx context.Context, batchID uuid.UUID) (bool, error) {
	args := m.Called(ctx, batchID)
	return args.Bool(0), args.Error(1)
}

func (m *mockBatchRepository) GetBatchHistory(ctx context.Context, batchID uuid.UUID, page, limit int) ([]response.BatchHistoryItemDTO, error) {
	args := m.Called(ctx, batchID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.BatchHistoryItemDTO), args.Error(1)
}

func (m *mockBatchRepository) GetBatchProducts(ctx context.Context, batchID uuid.UUID, req *request.GetBatchProductsRequest) (*response.GetBatchProductsResponse, error) {
	args := m.Called(ctx, batchID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.GetBatchProductsResponse), args.Error(1)
}

type mockPDFGenerator struct {
	mock.Mock
}

func (m *mockPDFGenerator) GenerateLabels(input qr.BatchPDFInput) ([]byte, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

type mockProductItemRepository struct {
	mock.Mock
}

func (m *mockProductItemRepository) FindByBatchID(ctx context.Context, batchID uuid.UUID) ([]*itemEntities.ProductItem, error) {
	args := m.Called(ctx, batchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*itemEntities.ProductItem), args.Error(1)
}

func (m *mockProductItemRepository) FindAllWithFilter(ctx context.Context, req *itemReq.GetProductItemListRequest) (*itemResp.ProductItemListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*itemResp.ProductItemListResponse), args.Error(1)
}

func (m *mockProductItemRepository) FindByItemCodeWithEvents(ctx context.Context, itemCode string) (*itemResp.ProductItemDetailDTO, error) {
	args := m.Called(ctx, itemCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*itemResp.ProductItemDetailDTO), args.Error(1)
}

func (m *mockProductItemRepository) Create(ctx context.Context, items []itemEntities.ProductItem) error {
	args := m.Called(ctx, items)
	return args.Error(0)
}

type mockProductVariantRepository struct {
	mock.Mock
}

func (m *mockProductVariantRepository) Create(ctx context.Context, variant *variantEntities.ProductVariant) error {
	args := m.Called(ctx, variant)
	return args.Error(0)
}
func (m *mockProductVariantRepository) Update(ctx context.Context, variant *variantEntities.ProductVariant) error {
	args := m.Called(ctx, variant)
	return args.Error(0)
}
func (m *mockProductVariantRepository) FindByID(ctx context.Context, id uuid.UUID) (*variantEntities.ProductVariant, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*variantEntities.ProductVariant), args.Error(1)
}
func (m *mockProductVariantRepository) FindByProductID(ctx context.Context, productID uuid.UUID) ([]variantEntities.ProductVariant, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).([]variantEntities.ProductVariant), args.Error(1)
}
func (m *mockProductVariantRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockProductVariantRepository) ExistsBySKU(ctx context.Context, sku string) (bool, error) {
	args := m.Called(ctx, sku)
	return args.Bool(0), args.Error(1)
}
func (m *mockProductVariantRepository) ExistsByBarcode(ctx context.Context, barcode string) (bool, error) {
	args := m.Called(ctx, barcode)
	return args.Bool(0), args.Error(1)
}
func (m *mockProductVariantRepository) ExistsBySKUExcludeID(ctx context.Context, sku string, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, sku, id)
	return args.Bool(0), args.Error(1)
}
func (m *mockProductVariantRepository) ExistsByBarcodeExcludeID(ctx context.Context, barcode string, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, barcode, id)
	return args.Bool(0), args.Error(1)
}
func (m *mockProductVariantRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

type mockAuditLogService struct {
	mock.Mock
}

func (m *mockAuditLogService) Log(ctx context.Context, userID *uuid.UUID, action string, entity string, entityID uuid.UUID, oldData any, newData any) error {
	args := m.Called(ctx, userID, action, entity, entityID, oldData, newData)
	return args.Error(0)
}
func (m *mockAuditLogService) LogCreate(ctx context.Context, userID *uuid.UUID, entity string, entityID uuid.UUID, newData any) error {
	args := m.Called(userID, entity, entityID, newData)
	return args.Error(0)
}
func (m *mockAuditLogService) LogUpdate(ctx context.Context, userID *uuid.UUID, entity string, entityID uuid.UUID, oldData any, newData any) error {
	args := m.Called(userID, entity, entityID, oldData, newData)
	return args.Error(0)
}
func (m *mockAuditLogService) LogDelete(ctx context.Context, userID *uuid.UUID, entity string, entityID uuid.UUID, oldData any) error {
	args := m.Called(userID, entity, entityID, oldData)
	return args.Error(0)
}

type mockRabbitMQPublisher struct {
	mock.Mock
}

func (m *mockRabbitMQPublisher) Publish(ctx context.Context, routingKey string, body []byte) error {
	args := m.Called(ctx, routingKey, body)
	return args.Error(0)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestGetBatchList_NonAdmin_Forbidden(t *testing.T) {
	t.Parallel()

	repo := new(mockBatchRepository)
	pdf := new(mockPDFGenerator)
	itemRepo := new(mockProductItemRepository)
	variantRepo := new(mockProductVariantRepository)
	audit := new(mockAuditLogService)

	svc := services.NewbatchService(repo, pdf, itemRepo, variantRepo, nil, audit)

	req := &request.GetBatchListRequest{
		Status: "DRAFT",
	}

	_, err := svc.GetBatchList(context.Background(), req, "CUSTOMER")

	assert.Error(t, err)
	var appErr *apperror.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.CodeForbidden, appErr.Code)
}

func TestSearchBatches_KeywordTooLong(t *testing.T) {
	t.Parallel()

	repo := new(mockBatchRepository)
	pdf := new(mockPDFGenerator)
	itemRepo := new(mockProductItemRepository)
	variantRepo := new(mockProductVariantRepository)
	audit := new(mockAuditLogService)

	svc := services.NewbatchService(repo, pdf, itemRepo, variantRepo, nil, audit)

	longKeyword := ""
	for i := 0; i < 105; i++ {
		longKeyword += "a"
	}

	req := &request.SearchBatchRequest{
		Keyword: longKeyword,
	}

	_, err := svc.SearchBatches(context.Background(), req)

	assert.Error(t, err)
	var appErr *apperror.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.CodeBadRequest, appErr.Code)
}

func TestCreateBatch_InvalidManufactureDate(t *testing.T) {
	t.Parallel()

	repo := new(mockBatchRepository)
	pdf := new(mockPDFGenerator)
	itemRepo := new(mockProductItemRepository)
	variantRepo := new(mockProductVariantRepository)
	audit := new(mockAuditLogService)

	svc := services.NewbatchService(repo, pdf, itemRepo, variantRepo, nil, audit)

	variantID := uuid.New()
	variantRepo.On("ExistsByID", mock.Anything, variantID).Return(true, nil).Once()

	future := time.Now().Add(24 * time.Hour)
	expiry := future.Add(24 * time.Hour)
	req := &request.CreateBatchRequest{
		VariantID:       variantID,
		Prefix:          "APL",
		ManufactureDate: &future,
		ExpiryDate:      &expiry,
		Quantity:        5,
	}

	_, err := svc.CreateBatch(context.Background(), req, uuid.New())

	assert.Error(t, err)
	var appErr *apperror.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.CodeValidation, appErr.Code)
	assert.Contains(t, err.Error(), "manufacture_date must not be in the future")
}

func TestCreateBatch_ExpiryBeforeManufacture(t *testing.T) {
	t.Parallel()

	repo := new(mockBatchRepository)
	pdf := new(mockPDFGenerator)
	itemRepo := new(mockProductItemRepository)
	variantRepo := new(mockProductVariantRepository)
	audit := new(mockAuditLogService)

	svc := services.NewbatchService(repo, pdf, itemRepo, variantRepo, nil, audit)

	variantID := uuid.New()
	variantRepo.On("ExistsByID", mock.Anything, variantID).Return(true, nil).Once()

	manufacture := time.Now().Add(-2 * time.Hour)
	expiry := time.Now().Add(-10 * time.Hour)

	req := &request.CreateBatchRequest{
		VariantID:       variantID,
		Prefix:          "APL",
		ManufactureDate: &manufacture,
		ExpiryDate:      &expiry,
		Quantity:        5,
	}

	_, err := svc.CreateBatch(context.Background(), req, uuid.New())

	assert.Error(t, err)
	var appErr *apperror.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.CodeValidation, appErr.Code)
	assert.Contains(t, err.Error(), "expiry_date must be greater than or equal to manufacture_date")
}

func TestUpdateBatchStatus_NonExisting(t *testing.T) {
	t.Parallel()

	repo := new(mockBatchRepository)
	pdf := new(mockPDFGenerator)
	itemRepo := new(mockProductItemRepository)
	variantRepo := new(mockProductVariantRepository)
	audit := new(mockAuditLogService)

	svc := services.NewbatchService(repo, pdf, itemRepo, variantRepo, nil, audit)

	batchID := uuid.New()
	repo.On("FindByID", mock.Anything, batchID).Return((*batchEntities.Batch)(nil), apperror.NewNotFound("Batch")).Once()

	req := &request.UpdateBatchStatusRequest{
		Status: "ACTIVE",
	}

	_, err := svc.UpdateBatchStatus(context.Background(), batchID, req, nil)

	assert.Error(t, err)
	var appErr *apperror.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.CodeNotFound, appErr.Code)
}

func TestDeleteBatch_WithLinkedItems(t *testing.T) {
	t.Parallel()

	repo := new(mockBatchRepository)
	pdf := new(mockPDFGenerator)
	itemRepo := new(mockProductItemRepository)
	variantRepo := new(mockProductVariantRepository)
	audit := new(mockAuditLogService)

	svc := services.NewbatchService(repo, pdf, itemRepo, variantRepo, nil, audit)

	batchID := uuid.New()
	batch := &batchEntities.Batch{
		ID:        batchID,
		BatchCode: "APL-2026-0001",
		IsDeleted: false,
	}

	repo.On("FindByID", mock.Anything, batchID).Return(batch, nil).Once()
	repo.On("ExistsProductItems", mock.Anything, batchID).Return(true, nil).Once()

	err := svc.DeleteBatch(context.Background(), batchID, nil)

	assert.Error(t, err)
	var appErr *apperror.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperror.CodeBadRequest, appErr.Code)
	assert.Contains(t, err.Error(), "cannot delete batch: batch has linked product items")
}
