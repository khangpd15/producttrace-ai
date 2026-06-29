package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/domain"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ─── helpers ────────────────────────────────────────────────────────────────

func float64Ptr(v float64) *float64 { return &v }

// newValidCreateReq trả về một CreateLocationReq hợp lệ cho mọi test case.
func newValidCreateReq() *dto.CreateLocationReq {
	return &dto.CreateLocationReq{
		OwnerUserID:      "owner-uuid-001",
		Code:             "LOC-001",
		Name:             "Kho Hà Nội",
		Type:             domain.TypeWarehouse,
		Phone:            "0912345678",
		Email:            "hanoi@example.com",
		Address:          "123 Đường Láng",
		Ward:             "Láng Hạ",
		District:         "Đống Đa",
		City:             "Hà Nội",
		Latitude:         float64Ptr(21.0278),
		Longitude:        float64Ptr(105.8342),
		OpeningHoursJSON: domain.OpeningHours{
			"monday": {Open: "08:00", Close: "18:00"},
		},
	}
}

// ─── TC-01: Happy path — tạo thành công với đầy đủ thông tin ─────────────────

func TestCreateLocation_HappyPath_Success(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()

	// GetByCode trả về (nil, nil) → code chưa tồn tại
	mockRepo.On("GetByCode", ctx, req.Code).Return(nil, nil)
	// Create thành công
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Location")).Return(nil)

	resp, err := svc.CreateLocation(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify response mapping
	assert.Equal(t, req.OwnerUserID, resp.OwnerUserID)
	assert.Equal(t, req.Code, resp.Code)
	assert.Equal(t, req.Name, resp.Name)
	assert.Equal(t, req.Type, resp.Type)
	assert.Equal(t, req.Phone, resp.Phone)
	assert.Equal(t, req.Email, resp.Email)
	assert.Equal(t, req.Address, resp.Address)
	assert.Equal(t, req.Ward, resp.Ward)
	assert.Equal(t, req.District, resp.District)
	assert.Equal(t, req.City, resp.City)
	assert.Equal(t, *req.Latitude, resp.Latitude)
	assert.Equal(t, *req.Longitude, resp.Longitude)
	assert.Equal(t, req.OpeningHoursJSON, resp.OpeningHoursJSON)

	// Verify hardcoded fields
	assert.Equal(t, "Vietnam", resp.Country)
	assert.True(t, resp.IsActive)

	// Verify ID sinh ra không rỗng
	assert.NotEmpty(t, resp.ID)

	mockRepo.AssertExpectations(t)
}

// ─── TC-02: Happy path — tạo thành công với OpeningHoursJSON = nil ──────────

func TestCreateLocation_HappyPath_NilOpeningHours(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()
	req.OpeningHoursJSON = nil

	mockRepo.On("GetByCode", ctx, req.Code).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Location")).Return(nil)

	resp, err := svc.CreateLocation(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Nil(t, resp.OpeningHoursJSON)
	assert.NotEmpty(t, resp.ID)

	mockRepo.AssertExpectations(t)
}

// ─── TC-03: Code đã tồn tại → trả lỗi "already exists" ──────────────────────

func TestCreateLocation_DuplicateCode_ReturnsError(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()

	existingLoc := &domain.Location{
		ID:   "existing-uuid",
		Code: req.Code,
	}

	// GetByCode trả về existing record
	mockRepo.On("GetByCode", ctx, req.Code).Return(existingLoc, nil)

	resp, err := svc.CreateLocation(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)

	// Verify error message chứa code name và thông báo đúng
	assert.Contains(t, err.Error(), "already exists")
	assert.Contains(t, err.Error(), req.Code)

	// Verify Create KHÔNG được gọi khi code đã tồn tại
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	mockRepo.AssertExpectations(t)
}

// ─── TC-04: repo.Create thất bại → trả lỗi được wrap ────────────────────────

func TestCreateLocation_CreateFails_ReturnsWrappedError(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()

	dbErr := errors.New("database connection timeout")

	mockRepo.On("GetByCode", ctx, req.Code).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Location")).Return(dbErr)

	resp, err := svc.CreateLocation(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)

	// Verify error được wrap đúng cách
	assert.Contains(t, err.Error(), "locationService.CreateLocation")
	assert.True(t, errors.Is(err, dbErr), "error phải wrap dbErr gốc qua %%w")

	mockRepo.AssertExpectations(t)
}

// ─── TC-05: GetByCode trả về error (bị ignore) → vẫn tiếp tục Create ────────

func TestCreateLocation_GetByCodeReturnsError_CreateStillCalled(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()

	repoErr := errors.New("db timeout on GetByCode")

	// GetByCode trả về (nil, error) — service bỏ qua error này
	mockRepo.On("GetByCode", ctx, req.Code).Return(nil, repoErr)
	// Create vẫn được gọi
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Location")).Return(nil)

	resp, err := svc.CreateLocation(ctx, req)

	// Service KHÔNG trả error từ GetByCode
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify Create được gọi dù GetByCode lỗi
	mockRepo.AssertCalled(t, "Create", ctx, mock.AnythingOfType("*domain.Location"))
	mockRepo.AssertExpectations(t)
}

// ─── TC-06: GetByCode trả về (*Location, error) → ưu tiên check existing != nil ─

func TestCreateLocation_GetByCodeReturnsBothLocAndError_TreatsAsExisting(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()

	existingLoc := &domain.Location{ID: "some-id", Code: req.Code}
	repoErr := errors.New("some partial error")

	// GetByCode trả về cả (*Location, error) — service chỉ check existing != nil
	mockRepo.On("GetByCode", ctx, req.Code).Return(existingLoc, repoErr)

	resp, err := svc.CreateLocation(ctx, req)

	// Vì existing != nil, service trả lỗi "already exists"
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "already exists")

	// Verify Create không được gọi
	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	mockRepo.AssertExpectations(t)
}

// ─── TC-07 / TC-08 / TC-09: Verify hardcoded fields trong domain object ──────

func TestCreateLocation_HardcodedFields_CountryIsActiveIsDeleted(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()

	var capturedLoc *domain.Location

	mockRepo.On("GetByCode", ctx, req.Code).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Location")).
		Run(func(args mock.Arguments) {
			capturedLoc = args.Get(1).(*domain.Location)
		}).
		Return(nil)

	_, err := svc.CreateLocation(ctx, req)
	require.NoError(t, err)

	require.NotNil(t, capturedLoc, "domain.Location phải được truyền vào Create")

	// TC-07: Country hardcoded = "Vietnam"
	assert.Equal(t, "Vietnam", capturedLoc.Country,
		"Country phải luôn là 'Vietnam' bất kể request")

	// TC-08: IsActive hardcoded = true
	assert.True(t, capturedLoc.IsActive,
		"IsActive phải luôn là true khi tạo mới")

	// TC-09: IsDeleted hardcoded = false
	assert.False(t, capturedLoc.IsDeleted,
		"IsDeleted phải luôn là false khi tạo mới")

	mockRepo.AssertExpectations(t)
}

// ─── TC-10: Latitude và Longitude được dereference đúng từ pointer ───────────

func TestCreateLocation_LatLng_DereferencedCorrectly(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()

	lat := 10.7769
	lng := 106.7009
	req.Latitude = &lat
	req.Longitude = &lng

	var capturedLoc *domain.Location

	mockRepo.On("GetByCode", ctx, req.Code).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Location")).
		Run(func(args mock.Arguments) {
			capturedLoc = args.Get(1).(*domain.Location)
		}).
		Return(nil)

	resp, err := svc.CreateLocation(ctx, req)
	require.NoError(t, err)

	// Verify trong domain object
	assert.Equal(t, lat, capturedLoc.Latitude)
	assert.Equal(t, lng, capturedLoc.Longitude)

	// Verify trong response
	assert.Equal(t, lat, resp.Latitude)
	assert.Equal(t, lng, resp.Longitude)

	mockRepo.AssertExpectations(t)
}

// ─── TC-11: ID sinh ra là UUID hợp lệ (non-empty, valid format) ──────────────

func TestCreateLocation_GeneratedID_IsValidUUID(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()

	var capturedLoc *domain.Location

	mockRepo.On("GetByCode", ctx, req.Code).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Location")).
		Run(func(args mock.Arguments) {
			capturedLoc = args.Get(1).(*domain.Location)
		}).
		Return(nil)

	resp, err := svc.CreateLocation(ctx, req)
	require.NoError(t, err)

	// Verify ID không rỗng
	assert.NotEmpty(t, capturedLoc.ID)
	// Verify UUID format: 8-4-4-4-12 (36 ký tự với dấu gạch)
	assert.Len(t, capturedLoc.ID, 36)
	parts := strings.Split(capturedLoc.ID, "-")
	assert.Len(t, parts, 5, "UUID phải có 5 phần phân cách bằng dấu '-'")

	// Verify ID đồng nhất giữa domain object và response
	assert.Equal(t, capturedLoc.ID, resp.ID)

	mockRepo.AssertExpectations(t)
}

// ─── TC-12: Toàn bộ fields mapping từ req → domain.Location ─────────────────

func TestCreateLocation_FieldMapping_ReqToDomain(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()

	var capturedLoc *domain.Location

	mockRepo.On("GetByCode", ctx, req.Code).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Location")).
		Run(func(args mock.Arguments) {
			capturedLoc = args.Get(1).(*domain.Location)
		}).
		Return(nil)

	_, err := svc.CreateLocation(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, capturedLoc)

	assert.Equal(t, req.OwnerUserID, capturedLoc.OwnerUserID)
	assert.Equal(t, req.Code, capturedLoc.Code)
	assert.Equal(t, req.Name, capturedLoc.Name)
	assert.Equal(t, req.Type, capturedLoc.Type)
	assert.Equal(t, req.Phone, capturedLoc.Phone)
	assert.Equal(t, req.Email, capturedLoc.Email)
	assert.Equal(t, req.Address, capturedLoc.Address)
	assert.Equal(t, req.Ward, capturedLoc.Ward)
	assert.Equal(t, req.District, capturedLoc.District)
	assert.Equal(t, req.City, capturedLoc.City)
	assert.Equal(t, req.OpeningHoursJSON, capturedLoc.OpeningHoursJSON)

	mockRepo.AssertExpectations(t)
}

// ─── TC-13: Response mapping — toResponse ánh xạ đúng từ domain sau Create ──

func TestCreateLocation_ResponseMapping_AllFieldsCorrect(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()

	// Giả lập Create có side-effect set CreatedAt/UpdatedAt (trong GORM thật)
	// Trong test, ta kiểm tra toResponse map đúng các field từ domain object
	var capturedLoc *domain.Location

	mockRepo.On("GetByCode", ctx, req.Code).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Location")).
		Run(func(args mock.Arguments) {
			capturedLoc = args.Get(1).(*domain.Location)
			// Simulate GORM setting timestamps
			capturedLoc.CreatedAt = time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC)
			capturedLoc.UpdatedAt = time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC)
		}).
		Return(nil)

	resp, err := svc.CreateLocation(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify toResponse ánh xạ đúng tất cả fields
	assert.Equal(t, capturedLoc.ID, resp.ID)
	assert.Equal(t, capturedLoc.OwnerUserID, resp.OwnerUserID)
	assert.Equal(t, capturedLoc.Code, resp.Code)
	assert.Equal(t, capturedLoc.Name, resp.Name)
	assert.Equal(t, capturedLoc.Type, resp.Type)
	assert.Equal(t, capturedLoc.Phone, resp.Phone)
	assert.Equal(t, capturedLoc.Email, resp.Email)
	assert.Equal(t, capturedLoc.Address, resp.Address)
	assert.Equal(t, capturedLoc.Ward, resp.Ward)
	assert.Equal(t, capturedLoc.District, resp.District)
	assert.Equal(t, capturedLoc.City, resp.City)
	assert.Equal(t, capturedLoc.Country, resp.Country)
	assert.Equal(t, capturedLoc.Latitude, resp.Latitude)
	assert.Equal(t, capturedLoc.Longitude, resp.Longitude)
	assert.Equal(t, capturedLoc.OpeningHoursJSON, resp.OpeningHoursJSON)
	assert.Equal(t, capturedLoc.IsActive, resp.IsActive)
	assert.Equal(t, capturedLoc.CreatedAt, resp.CreatedAt)
	assert.Equal(t, capturedLoc.UpdatedAt, resp.UpdatedAt)

	mockRepo.AssertExpectations(t)
}

// ─── TC-14: CreatedAt và UpdatedAt được map vào response ─────────────────────

func TestCreateLocation_ResponseMapping_TimestampsAreMapped(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()

	expectedTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	mockRepo.On("GetByCode", ctx, req.Code).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Location")).
		Run(func(args mock.Arguments) {
			loc := args.Get(1).(*domain.Location)
			loc.CreatedAt = expectedTime
			loc.UpdatedAt = expectedTime
		}).
		Return(nil)

	resp, err := svc.CreateLocation(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, expectedTime, resp.CreatedAt)
	assert.Equal(t, expectedTime, resp.UpdatedAt)

	mockRepo.AssertExpectations(t)
}

// ─── TC-15: GetByCode được gọi đúng 1 lần với req.Code ──────────────────────

func TestCreateLocation_RepoInteraction_GetByCodeCalledOnce(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()

	mockRepo.On("GetByCode", ctx, req.Code).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Location")).Return(nil)

	_, err := svc.CreateLocation(ctx, req)
	require.NoError(t, err)

	// Verify GetByCode được gọi đúng 1 lần với đúng code
	mockRepo.AssertNumberOfCalls(t, "GetByCode", 1)
	mockRepo.AssertCalled(t, "GetByCode", ctx, req.Code)
	mockRepo.AssertExpectations(t)
}

// ─── TC-16: Khi code trùng, Create KHÔNG được gọi ────────────────────────────

func TestCreateLocation_RepoInteraction_CreateNotCalledWhenDuplicate(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()

	existingLoc := &domain.Location{ID: "existing-id", Code: req.Code}
	mockRepo.On("GetByCode", ctx, req.Code).Return(existingLoc, nil)

	resp, err := svc.CreateLocation(ctx, req)

	require.Error(t, err)
	assert.Nil(t, resp)

	mockRepo.AssertNumberOfCalls(t, "GetByCode", 1)
	mockRepo.AssertNumberOfCalls(t, "Create", 0)
	mockRepo.AssertExpectations(t)
}

// ─── TC-17: Khi thành công, Create được gọi đúng 1 lần với domain object hợp lệ

func TestCreateLocation_RepoInteraction_CreateCalledOnceWithValidDomainObject(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()
	req := newValidCreateReq()

	var capturedLoc *domain.Location

	mockRepo.On("GetByCode", ctx, req.Code).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Location")).
		Run(func(args mock.Arguments) {
			capturedLoc = args.Get(1).(*domain.Location)
		}).
		Return(nil)

	_, err := svc.CreateLocation(ctx, req)
	require.NoError(t, err)

	// Verify Create được gọi đúng 1 lần
	mockRepo.AssertNumberOfCalls(t, "Create", 1)

	// Verify object truyền vào Create là hợp lệ
	require.NotNil(t, capturedLoc)
	assert.NotEmpty(t, capturedLoc.ID)
	assert.Equal(t, req.Code, capturedLoc.Code)
	assert.Equal(t, "Vietnam", capturedLoc.Country)
	assert.True(t, capturedLoc.IsActive)
	assert.False(t, capturedLoc.IsDeleted)

	mockRepo.AssertExpectations(t)
}

// ─── Bonus: Verify mỗi lần gọi sinh ra ID khác nhau (UUID uniqueness) ────────

func TestCreateLocation_EachCall_GeneratesUniqueID(t *testing.T) {
	mockRepo := new(mocks.MockLocationRepository)
	svc := NewLocationService(mockRepo)
	ctx := context.Background()

	var capturedLocs []*domain.Location

	// Cấu hình mock chung cho tất cả các cuộc gọi trong test case này
	mockRepo.On("GetByCode", ctx, mock.Anything).Return(nil, nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Location")).
		Run(func(args mock.Arguments) {
			loc := args.Get(1).(*domain.Location)
			capturedLocs = append(capturedLocs, loc)
		}).
		Return(nil)

	for i := 0; i < 3; i++ {
		req := newValidCreateReq()
		req.Code = "LOC-00" + string(rune('1'+i)) // LOC-001, LOC-002, LOC-003

		_, err := svc.CreateLocation(ctx, req)
		require.NoError(t, err)
	}

	require.Len(t, capturedLocs, 3)

	// Verify 3 ID đều khác nhau
	assert.NotEmpty(t, capturedLocs[0].ID)
	assert.NotEmpty(t, capturedLocs[1].ID)
	assert.NotEmpty(t, capturedLocs[2].ID)
	assert.NotEqual(t, capturedLocs[0].ID, capturedLocs[1].ID)
	assert.NotEqual(t, capturedLocs[1].ID, capturedLocs[2].ID)
	assert.NotEqual(t, capturedLocs[0].ID, capturedLocs[2].ID)

	mockRepo.AssertExpectations(t)
}

