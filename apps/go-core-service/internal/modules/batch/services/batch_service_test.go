package services

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

// mockBatchRepository implement repositories.BatchRepository để dùng trong test.
// Dùng func field để từng test case tự quyết định hành vi trả về.
type mockBatchRepository struct {
	findAllFn    func(ctx context.Context) ([]*response.BatchListResponse, error)
	findByCodeFn func(ctx context.Context, batchCode string) (*response.BatchDetailResponse, error)
	createFn     func(ctx context.Context, req *request.CreateBatchRequest) (*response.BatchCreateResponse, error)
}

func (m *mockBatchRepository) FindAll(ctx context.Context) ([]*response.BatchListResponse, error) {
	return m.findAllFn(ctx)
}

func (m *mockBatchRepository) FindByCode(ctx context.Context, batchCode string) (*response.BatchDetailResponse, error) {
	return m.findByCodeFn(ctx, batchCode)
}

func (m *mockBatchRepository) Create(ctx context.Context, req *request.CreateBatchRequest) (*response.BatchCreateResponse, error) {
	return m.createFn(ctx, req)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func ptr[T any](v T) *T { return &v }

func newService(repo *mockBatchRepository) BatchService {
	return NewbatchService(repo)
}

// ---------------------------------------------------------------------------
// GetBatchList
// ---------------------------------------------------------------------------

func TestGetBatchList_Success(t *testing.T) {
	now := time.Now()
	productID := uuid.New()

	want := []*response.BatchListResponse{
		{BatchCode: "LOT-001", ProductID: productID, Status: "ACTIVE", CreatedAt: now, ExpiryDate: nil},
		{BatchCode: "LOT-002", ProductID: productID, Status: "ACTIVE", CreatedAt: now, ExpiryDate: ptr(now.Add(24 * time.Hour))},
	}

	mock := &mockBatchRepository{
		findAllFn: func(ctx context.Context) ([]*response.BatchListResponse, error) {
			return want, nil
		},
	}

	svc := newService(mock)
	got, err := svc.GetBatchList(context.Background())

	require.NoError(t, err)
	require.Len(t, got, len(want))
	assert.Equal(t, want[0].BatchCode, got[0].BatchCode)
	assert.Equal(t, want[1].BatchCode, got[1].BatchCode)
	assert.Nil(t, got[0].ExpiryDate)
	assert.NotNil(t, got[1].ExpiryDate)
}

func TestGetBatchList_EmptyResult(t *testing.T) {
	mock := &mockBatchRepository{
		findAllFn: func(ctx context.Context) ([]*response.BatchListResponse, error) {
			return []*response.BatchListResponse{}, nil
		},
	}

	svc := newService(mock)
	got, err := svc.GetBatchList(context.Background())

	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGetBatchList_RepoError(t *testing.T) {
	repoErr := apperror.NewInternal("database error")

	mock := &mockBatchRepository{
		findAllFn: func(ctx context.Context) ([]*response.BatchListResponse, error) {
			return nil, repoErr
		},
	}

	svc := newService(mock)
	got, err := svc.GetBatchList(context.Background())

	require.Error(t, err)
	assert.Nil(t, got)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeInternal, appErr.Code)
}

// ---------------------------------------------------------------------------
// GetBatchDetail
// ---------------------------------------------------------------------------

func buildBatchDetail(batchCode string) *response.BatchDetailResponse {
	now := time.Now()
	return &response.BatchDetailResponse{
		ID:               uuid.New(),
		BatchCode:        batchCode,
		ManufactureDate:  ptr(now.Add(-30 * 24 * time.Hour)),
		ExpiryDate:       ptr(now.Add(365 * 24 * time.Hour)),
		ImportedAt:       ptr(now.Add(-25 * 24 * time.Hour)),
		ManufacturerName: ptr("Nhà máy ABC"),
		SupplierName:     ptr("Nhà cung cấp XYZ"),
		OriginCountry:    ptr("Vietnam"),
		ProductionPlace:  ptr("Hà Nội"),
		Quantity:         5000,
		Status:           "ACTIVE",
		CreatedAt:        now,
		UpdatedAt:        now,
		Variant: response.BatchDetailVariantResponse{
			VariantID: uuid.New(),
			SKU:       "IP16PM-256-TITAN",
			Name:      "iPhone 16 Pro Max 256GB Titan",
			Barcode:   ptr("8931234567890"),
		},
		Product: response.BatchDetailProductResponse{
			ProductID:   uuid.New(),
			ProductName: "iPhone 16 Pro Max",
		},
	}
}

func TestGetBatchDetail_Success(t *testing.T) {
	const batchCode = "LOT-2026-06A"
	want := buildBatchDetail(batchCode)

	var capturedCode string
	mock := &mockBatchRepository{
		findByCodeFn: func(ctx context.Context, code string) (*response.BatchDetailResponse, error) {
			capturedCode = code
			return want, nil
		},
	}

	svc := newService(mock)
	got, err := svc.GetBatchDetail(context.Background(), batchCode)

	require.NoError(t, err)
	require.NotNil(t, got)

	// Kiểm tra batch_code được forward đúng xuống repo
	assert.Equal(t, batchCode, capturedCode, "batch_code phải được forward chính xác xuống repository")

	assert.Equal(t, want.BatchCode, got.BatchCode)
	assert.Equal(t, want.Status, got.Status)
	assert.Equal(t, want.Quantity, got.Quantity)
	assert.Equal(t, want.Variant.SKU, got.Variant.SKU)
	assert.Equal(t, want.Variant.Name, got.Variant.Name)
	assert.Equal(t, want.Product.ProductName, got.Product.ProductName)
	assert.NotNil(t, got.ManufactureDate)
	assert.NotNil(t, got.ExpiryDate)
	assert.NotNil(t, got.Variant.Barcode)
}

func TestGetBatchDetail_NullableFieldsHandled(t *testing.T) {
	const batchCode = "LOT-NULL-FIELDS"
	want := &response.BatchDetailResponse{
		ID:               uuid.New(),
		BatchCode:        batchCode,
		ManufactureDate:  nil,
		ExpiryDate:       nil,
		ImportedAt:       nil,
		ManufacturerName: nil,
		SupplierName:     nil,
		OriginCountry:    nil,
		ProductionPlace:  nil,
		Quantity:         0,
		Status:           "ACTIVE",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Variant: response.BatchDetailVariantResponse{
			VariantID: uuid.New(),
			SKU:       "SKU-001",
			Name:      "Variant A",
			Barcode:   nil,
		},
		Product: response.BatchDetailProductResponse{
			ProductID:   uuid.New(),
			ProductName: "Product A",
		},
	}

	mock := &mockBatchRepository{
		findByCodeFn: func(ctx context.Context, code string) (*response.BatchDetailResponse, error) {
			return want, nil
		},
	}

	svc := newService(mock)
	got, err := svc.GetBatchDetail(context.Background(), batchCode)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Nil(t, got.ManufactureDate)
	assert.Nil(t, got.ExpiryDate)
	assert.Nil(t, got.ImportedAt)
	assert.Nil(t, got.ManufacturerName)
	assert.Nil(t, got.SupplierName)
	assert.Nil(t, got.OriginCountry)
	assert.Nil(t, got.ProductionPlace)
	assert.Nil(t, got.Variant.Barcode)
	assert.Equal(t, 0, got.Quantity)
}

func TestGetBatchDetail_NotFound(t *testing.T) {
	notFoundErr := apperror.WrapDBError(sql.ErrNoRows, "batch")

	mock := &mockBatchRepository{
		findByCodeFn: func(ctx context.Context, code string) (*response.BatchDetailResponse, error) {
			return nil, notFoundErr
		},
	}

	svc := newService(mock)
	got, err := svc.GetBatchDetail(context.Background(), "BATCH-NOT-EXIST")

	require.Error(t, err)
	assert.Nil(t, got)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeNotFound, appErr.Code)
}

func TestGetBatchDetail_InternalError(t *testing.T) {
	dbErr := apperror.NewInternal("database error")

	mock := &mockBatchRepository{
		findByCodeFn: func(ctx context.Context, code string) (*response.BatchDetailResponse, error) {
			return nil, dbErr
		},
	}

	svc := newService(mock)
	got, err := svc.GetBatchDetail(context.Background(), "LOT-ANY")

	require.Error(t, err)
	assert.Nil(t, got)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeInternal, appErr.Code)
}

// ---------------------------------------------------------------------------
// CreateBatch
// ---------------------------------------------------------------------------

func buildCreateRequest() *request.CreateBatchRequest {
	now := time.Now()
	variantID := uuid.New()
	return &request.CreateBatchRequest{
		VariantID:        variantID,
		Prefix:           "APL",
		ManufactureDate:  ptr(now.Add(-30 * 24 * time.Hour)),
		ExpiryDate:       ptr(now.Add(365 * 24 * time.Hour)),
		ImportedAt:       ptr(now),
		ManufacturerName: ptr("Nhà máy ABC"),
		SupplierName:     ptr("Nhà cung cấp XYZ"),
		OriginCountry:    ptr("Vietnam"),
		ProductionPlace:  ptr("Hà Nội"),
		Quantity:         5000,
	}
}

func TestCreateBatch_Success(t *testing.T) {
	req := buildCreateRequest()

	want := &response.BatchCreateResponse{
		ID:        uuid.New(),
		BatchCode: "APL-2026-0001", // batch code do repo sinh ra
		VariantID: req.VariantID,
		Quantity:  req.Quantity,
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
	}

	var capturedReq *request.CreateBatchRequest
	mock := &mockBatchRepository{
		createFn: func(ctx context.Context, r *request.CreateBatchRequest) (*response.BatchCreateResponse, error) {
			capturedReq = r
			return want, nil
		},
	}

	svc := newService(mock)
	got, err := svc.CreateBatch(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, got)

	// Prefix phải đã được normalize (uppercase) trước khi forward xuống repo
	assert.Equal(t, "APL", capturedReq.Prefix)
	assert.Equal(t, req.VariantID, capturedReq.VariantID)
	assert.Equal(t, req.Quantity, capturedReq.Quantity)

	assert.Equal(t, want.BatchCode, got.BatchCode)
	assert.Equal(t, want.VariantID, got.VariantID)
	assert.Equal(t, want.Quantity, got.Quantity)
	assert.Equal(t, "ACTIVE", got.Status)
	assert.NotEqual(t, uuid.Nil, got.ID)
}

func TestCreateBatch_PrefixNormalization_LowerCase(t *testing.T) {
	// "apl" → service normalize → "APL" trước khi gọi repo
	req := buildCreateRequest()
	req.Prefix = "apl"

	var capturedPrefix string
	mock := &mockBatchRepository{
		createFn: func(ctx context.Context, r *request.CreateBatchRequest) (*response.BatchCreateResponse, error) {
			capturedPrefix = r.Prefix
			return &response.BatchCreateResponse{ID: uuid.New(), BatchCode: "APL-2026-0001", Status: "ACTIVE"}, nil
		},
	}

	svc := newService(mock)
	_, err := svc.CreateBatch(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "APL", capturedPrefix, "prefix phải được chuyển thành uppercase")
}

func TestCreateBatch_PrefixNormalization_MixedCaseWithSpaces(t *testing.T) {
	// " Sam " → service normalize → "SAM"
	req := buildCreateRequest()
	req.Prefix = " Sam "

	var capturedPrefix string
	mock := &mockBatchRepository{
		createFn: func(ctx context.Context, r *request.CreateBatchRequest) (*response.BatchCreateResponse, error) {
			capturedPrefix = r.Prefix
			return &response.BatchCreateResponse{ID: uuid.New(), BatchCode: "SAM-2026-0001", Status: "ACTIVE"}, nil
		},
	}

	svc := newService(mock)
	_, err := svc.CreateBatch(context.Background(), req)

	require.NoError(t, err)
	assert.Equal(t, "SAM", capturedPrefix, "prefix phải được trim và chuyển uppercase")
}

func TestCreateBatch_ExpiryBeforeManufacture_ValidationError(t *testing.T) {
	now := time.Now()
	req := buildCreateRequest()
	// Đặt expiry_date trước manufacture_date — vi phạm business rule
	req.ManufactureDate = ptr(now.Add(10 * 24 * time.Hour))
	req.ExpiryDate = ptr(now.Add(5 * 24 * time.Hour))

	mock := &mockBatchRepository{
		// createFn KHÔNG được gọi trong trường hợp này
		createFn: func(ctx context.Context, r *request.CreateBatchRequest) (*response.BatchCreateResponse, error) {
			t.Fatal("repo.Create không nên được gọi khi validation thất bại")
			return nil, nil
		},
	}

	svc := newService(mock)
	got, err := svc.CreateBatch(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, got)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeValidation, appErr.Code)
}

func TestCreateBatch_OnlyExpiryDate_NoValidationError(t *testing.T) {
	// Nếu chỉ có expiry_date mà không có manufacture_date → không lỗi validation
	req := buildCreateRequest()
	req.ManufactureDate = nil
	req.ExpiryDate = ptr(time.Now().Add(365 * 24 * time.Hour))

	want := &response.BatchCreateResponse{
		ID: uuid.New(), BatchCode: "APL-2026-0001", Status: "ACTIVE",
	}
	mock := &mockBatchRepository{
		createFn: func(ctx context.Context, r *request.CreateBatchRequest) (*response.BatchCreateResponse, error) {
			return want, nil
		},
	}

	svc := newService(mock)
	got, err := svc.CreateBatch(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, got)
}

func TestCreateBatch_DifferentPrefixes_IndependentSequences(t *testing.T) {
	// SAM và XMI là 2 prefix khác nhau → lock key khác nhau → chạy song song
	// Test đảm bảo service forward đúng prefix xuống repo
	prefixes := []string{"SAM", "XMI", "APL"}

	for _, prefix := range prefixes {
		prefix := prefix // capture loop var
		t.Run("prefix_"+prefix, func(t *testing.T) {
			req := buildCreateRequest()
			req.Prefix = prefix

			var capturedPrefix string
			mock := &mockBatchRepository{
				createFn: func(ctx context.Context, r *request.CreateBatchRequest) (*response.BatchCreateResponse, error) {
					capturedPrefix = r.Prefix
					return &response.BatchCreateResponse{
						ID: uuid.New(), BatchCode: prefix + "-2026-0001", Status: "ACTIVE",
					}, nil
				},
			}

			svc := newService(mock)
			_, err := svc.CreateBatch(context.Background(), req)

			require.NoError(t, err)
			assert.Equal(t, prefix, capturedPrefix)
		})
	}
}

func TestCreateBatch_ConflictError(t *testing.T) {
	req := buildCreateRequest()
	conflictErr := apperror.NewConflict("batch already exists")

	mock := &mockBatchRepository{
		createFn: func(ctx context.Context, r *request.CreateBatchRequest) (*response.BatchCreateResponse, error) {
			return nil, conflictErr
		},
	}

	svc := newService(mock)
	got, err := svc.CreateBatch(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, got)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeConflict, appErr.Code)
}

func TestCreateBatch_RepoError(t *testing.T) {
	req := buildCreateRequest()
	dbErr := apperror.NewInternal("database error")

	mock := &mockBatchRepository{
		createFn: func(ctx context.Context, r *request.CreateBatchRequest) (*response.BatchCreateResponse, error) {
			return nil, dbErr
		},
	}

	svc := newService(mock)
	got, err := svc.CreateBatch(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, got)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeInternal, appErr.Code)
}
