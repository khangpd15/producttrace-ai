package services

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/google/uuid"
// 	batchDTORequest "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/request"
// 	batchDTOResponse "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/response"
// 	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/dto/request"
// 	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/dto/response"
// 	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/entities"
// 	variantEntities "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/entities"
// 	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// // ---------------------------------------------------------------------------
// // Mock ProductItemRepository
// // ---------------------------------------------------------------------------

// type mockProductItemRepository struct {
// 	createFn      func(ctx context.Context, pi *entities.ProductItem) (*response.ProductItemCreateResponse, error)
// 	findByBatchFn func(ctx context.Context, batchID uuid.UUID) ([]*entities.ProductItem, error)
// }

// func (m *mockProductItemRepository) Create(ctx context.Context, pi *entities.ProductItem) (*response.ProductItemCreateResponse, error) {
// 	if m.createFn != nil {
// 		return m.createFn(ctx, pi)
// 	}
// 	batchIDCopy := pi.BatchID
// 	return &response.ProductItemCreateResponse{
// 		ID:                pi.ID,
// 		VariantID:         pi.VariantID,
// 		BatchID:           &batchIDCopy,
// 		ItemCode:          pi.ItemCode,
// 		SerialNumber:      pi.SerialNumber,
// 		VerificationToken: pi.VerificationToken,
// 		Status:            "ACTIVE",
// 		CreatedAt:         time.Now(),
// 	}, nil
// }

// func (m *mockProductItemRepository) FindByBatchID(ctx context.Context, batchID uuid.UUID) ([]*entities.ProductItem, error) {
// 	if m.findByBatchFn != nil {
// 		return m.findByBatchFn(ctx, batchID)
// 	}
// 	return nil, nil
// }

// // ---------------------------------------------------------------------------
// // Mock BatchRepository (chỉ cần ExistsByID cho ProductItem test)
// // ---------------------------------------------------------------------------

// type mockBatchRepo struct {
// 	existsByIDFn func(ctx context.Context, id uuid.UUID) (bool, error)
// }

// func (m *mockBatchRepo) FindAll(ctx context.Context) ([]*batchDTOResponse.BatchListResponse, error) {
// 	return nil, nil
// }

// func (m *mockBatchRepo) FindByCode(ctx context.Context, batchCode string) (*batchDTOResponse.BatchDetailResponse, error) {
// 	return nil, nil
// }

// func (m *mockBatchRepo) FindByBatchID(ctx context.Context, batchID uuid.UUID) (*batchDTOResponse.BatchDetailResponse, error) {
// 	return nil, nil
// }

// func (m *mockBatchRepo) Create(ctx context.Context, req *batchDTORequest.CreateBatchRequest) (*batchDTOResponse.BatchCreateResponse, error) {
// 	return nil, nil
// }

// func (m *mockBatchRepo) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
// 	if m.existsByIDFn != nil {
// 		return m.existsByIDFn(ctx, id)
// 	}
// 	return true, nil
// }

// // ---------------------------------------------------------------------------
// // Mock ProductVariantRepository
// // ---------------------------------------------------------------------------

// type mockVariantRepo struct {
// 	existsByIDFn func(ctx context.Context, id uuid.UUID) (bool, error)
// }

// func (m *mockVariantRepo) Create(ctx context.Context, variant *variantEntities.ProductVariant) error {
// 	return nil
// }

// func (m *mockVariantRepo) ExistsBySKU(ctx context.Context, sku string) (bool, error) {
// 	return false, nil
// }

// func (m *mockVariantRepo) ExistsByBarcode(ctx context.Context, barcode string) (bool, error) {
// 	return false, nil
// }

// func (m *mockVariantRepo) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
// 	if m.existsByIDFn != nil {
// 		return m.existsByIDFn(ctx, id)
// 	}
// 	return true, nil
// }

// // ---------------------------------------------------------------------------
// // Helpers
// // ---------------------------------------------------------------------------

// func buildCreateProductItemRequest(batchID uuid.UUID) *request.CreateProductItemRequest {
// 	return &request.CreateProductItemRequest{
// 		VariantID: uuid.New(),
// 		BatchID:   batchID,
// 	}
// }

// func newProductItemService(
// 	itemRepo *mockProductItemRepository,
// 	bRepo *mockBatchRepo,
// 	vRepo *mockVariantRepo,
// ) ProductItemService {
// 	return NewProductItemService(itemRepo, bRepo, vRepo)
// }

// // ---------------------------------------------------------------------------
// // CreateProductItem – Foreign Key Validation
// // ---------------------------------------------------------------------------

// func TestCreateProductItem_Success(t *testing.T) {
// 	// Cả batch_id và variant_id đều tồn tại → tạo thành công
// 	batchID := uuid.New()
// 	req := buildCreateProductItemRequest(batchID)

// 	itemRepo := &mockProductItemRepository{}
// 	bRepo := &mockBatchRepo{
// 		existsByIDFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
// 			assert.Equal(t, batchID, id)
// 			return true, nil
// 		},
// 	}
// 	vRepo := &mockVariantRepo{
// 		existsByIDFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
// 			assert.Equal(t, req.VariantID, id)
// 			return true, nil
// 		},
// 	}

// 	svc := newProductItemService(itemRepo, bRepo, vRepo)
// 	got, err := svc.CreateProductItem(context.Background(), req)

// 	require.NoError(t, err)
// 	require.NotNil(t, got)
// 	assert.Equal(t, req.VariantID, got.VariantID)
// 	require.NotNil(t, got.BatchID)
// 	assert.Equal(t, req.BatchID, *got.BatchID)
// 	assert.Equal(t, "ACTIVE", got.Status)
// }

// func TestCreateProductItem_NoBatchID_Success(t *testing.T) {
// 	// Không truyền batch_id (uuid.Nil) → bỏ qua check batch → tạo thành công
// 	req := buildCreateProductItemRequest(uuid.Nil)

// 	itemRepo := &mockProductItemRepository{}
// 	bRepo := &mockBatchRepo{
// 		existsByIDFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
// 			t.Fatal("batchRepo.ExistsByID không nên được gọi khi BatchID là uuid.Nil")
// 			return false, nil
// 		},
// 	}
// 	vRepo := &mockVariantRepo{}

// 	svc := newProductItemService(itemRepo, bRepo, vRepo)
// 	got, err := svc.CreateProductItem(context.Background(), req)

// 	require.NoError(t, err)
// 	require.NotNil(t, got)
// }

// func TestCreateProductItem_BatchNotFound_Error(t *testing.T) {
// 	// batch_id được cung cấp nhưng không tồn tại → 404
// 	batchID := uuid.New()
// 	req := buildCreateProductItemRequest(batchID)

// 	itemRepo := &mockProductItemRepository{
// 		createFn: func(ctx context.Context, pi *entities.ProductItem) (*response.ProductItemCreateResponse, error) {
// 			t.Fatal("repo.Create không nên được gọi khi batch không tồn tại")
// 			return nil, nil
// 		},
// 	}
// 	bRepo := &mockBatchRepo{
// 		existsByIDFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
// 			return false, nil // batch không tồn tại
// 		},
// 	}
// 	vRepo := &mockVariantRepo{}

// 	svc := newProductItemService(itemRepo, bRepo, vRepo)
// 	got, err := svc.CreateProductItem(context.Background(), req)

// 	require.Error(t, err)
// 	assert.Nil(t, got)

// 	var appErr *apperror.AppError
// 	require.ErrorAs(t, err, &appErr)
// 	assert.Equal(t, apperror.CodeNotFound, appErr.Code)
// 	assert.Contains(t, appErr.Message, "batch")
// }

// func TestCreateProductItem_VariantNotFound_Error(t *testing.T) {
// 	// variant_id không tồn tại → 404
// 	req := buildCreateProductItemRequest(uuid.New())

// 	itemRepo := &mockProductItemRepository{
// 		createFn: func(ctx context.Context, pi *entities.ProductItem) (*response.ProductItemCreateResponse, error) {
// 			t.Fatal("repo.Create không nên được gọi khi variant không tồn tại")
// 			return nil, nil
// 		},
// 	}
// 	bRepo := &mockBatchRepo{} // batch tồn tại (mặc định true)
// 	vRepo := &mockVariantRepo{
// 		existsByIDFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
// 			return false, nil // variant không tồn tại
// 		},
// 	}

// 	svc := newProductItemService(itemRepo, bRepo, vRepo)
// 	got, err := svc.CreateProductItem(context.Background(), req)

// 	require.Error(t, err)
// 	assert.Nil(t, got)

// 	var appErr *apperror.AppError
// 	require.ErrorAs(t, err, &appErr)
// 	assert.Equal(t, apperror.CodeNotFound, appErr.Code)
// 	assert.Contains(t, appErr.Message, "product_variant")
// }

// func TestCreateProductItem_BatchCheckInternalError(t *testing.T) {
// 	// Lỗi DB khi check batch → 500
// 	batchID := uuid.New()
// 	req := buildCreateProductItemRequest(batchID)
// 	dbErr := apperror.NewInternal("database error")

// 	itemRepo := &mockProductItemRepository{
// 		createFn: func(ctx context.Context, pi *entities.ProductItem) (*response.ProductItemCreateResponse, error) {
// 			t.Fatal("repo.Create không nên được gọi khi check batch lỗi DB")
// 			return nil, nil
// 		},
// 	}
// 	bRepo := &mockBatchRepo{
// 		existsByIDFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
// 			return false, dbErr
// 		},
// 	}
// 	vRepo := &mockVariantRepo{}

// 	svc := newProductItemService(itemRepo, bRepo, vRepo)
// 	got, err := svc.CreateProductItem(context.Background(), req)

// 	require.Error(t, err)
// 	assert.Nil(t, got)

// 	var appErr *apperror.AppError
// 	require.ErrorAs(t, err, &appErr)
// 	assert.Equal(t, apperror.CodeInternal, appErr.Code)
// }

// func TestCreateProductItem_VariantCheckInternalError(t *testing.T) {
// 	// Lỗi DB khi check variant → 500
// 	req := buildCreateProductItemRequest(uuid.New())
// 	dbErr := apperror.NewInternal("database error")

// 	itemRepo := &mockProductItemRepository{
// 		createFn: func(ctx context.Context, pi *entities.ProductItem) (*response.ProductItemCreateResponse, error) {
// 			t.Fatal("repo.Create không nên được gọi khi check variant lỗi DB")
// 			return nil, nil
// 		},
// 	}
// 	bRepo := &mockBatchRepo{} // batch tồn tại (mặc định true)
// 	vRepo := &mockVariantRepo{
// 		existsByIDFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
// 			return false, dbErr
// 		},
// 	}

// 	svc := newProductItemService(itemRepo, bRepo, vRepo)
// 	got, err := svc.CreateProductItem(context.Background(), req)

// 	require.Error(t, err)
// 	assert.Nil(t, got)

// 	var appErr *apperror.AppError
// 	require.ErrorAs(t, err, &appErr)
// 	assert.Equal(t, apperror.CodeInternal, appErr.Code)
// }
