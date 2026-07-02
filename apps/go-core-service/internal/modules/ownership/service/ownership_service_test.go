package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockOwnershipRepo struct {
	createOwnershipFn                   func(ctx context.Context, tx *gorm.DB, o *entity.Ownership) (*entity.Ownership, error)
	getOwnershipByProductItemIDFn       func(ctx context.Context, productItemID uuid.UUID) (*entity.Ownership, error)
	getOwnershipHistoryByProductItemIDFn func(ctx context.Context, productItemID uuid.UUID) ([]entity.Ownership, error)
}

func (m *mockOwnershipRepo) CreateOwnership(ctx context.Context, tx *gorm.DB, o *entity.Ownership) (*entity.Ownership, error) {
	return m.createOwnershipFn(ctx, tx, o)
}
func (m *mockOwnershipRepo) GetOwnershipByProductItemID(ctx context.Context, productItemID uuid.UUID) (*entity.Ownership, error) {
	return m.getOwnershipByProductItemIDFn(ctx, productItemID)
}
func (m *mockOwnershipRepo) GetOwnershipHistoryByProductItemID(ctx context.Context, productItemID uuid.UUID) ([]entity.Ownership, error) {
	return m.getOwnershipHistoryByProductItemIDFn(ctx, productItemID)
}

type mockProductPort struct {
	findProductByQRFn                func(ctx context.Context, qr string) (uuid.UUID, error)
	validateProductOwnershipStatusFn func(ctx context.Context, productID uuid.UUID) error
	updateOwnershipStatusFn          func(ctx context.Context, productID uuid.UUID, status string) error
}

func (m *mockProductPort) FindProductByQR(ctx context.Context, qr string) (uuid.UUID, error) {
	return m.findProductByQRFn(ctx, qr)
}
func (m *mockProductPort) ValidateProductOwnershipStatus(ctx context.Context, id uuid.UUID) error {
	return m.validateProductOwnershipStatusFn(ctx, id)
}
func (m *mockProductPort) UpdateOwnershipStatus(ctx context.Context, id uuid.UUID, status string) error {
	return m.updateOwnershipStatusFn(ctx, id, status)
}
func (m *mockProductPort) GetProductItemDetail(_ context.Context, _ uuid.UUID) (string, string, error) {
	return "Demo Product", "SKU-001", nil
}

type mockEmailPort struct {
	requestOTPFn  func(ctx context.Context, email, productIDStr string) error
	validateOTPFn func(ctx context.Context, email, otp string) (bool, error)
}

func (m *mockEmailPort) RequestOTP(ctx context.Context, email, productIDStr string) error {
	return m.requestOTPFn(ctx, email, productIDStr)
}
func (m *mockEmailPort) ValidateOTP(ctx context.Context, email, otp string) (bool, error) {
	return m.validateOTPFn(ctx, email, otp)
}

type mockUserProvider struct {
	getUserEmailByIDFn func(ctx context.Context, userID uuid.UUID) (string, string, string, error)
	ensureUserExistsFn func(ctx context.Context, email, fullName, phone string) (uuid.UUID, error)
}

func (m *mockUserProvider) GetUserEmailByID(ctx context.Context, userID uuid.UUID) (string, string, string, error) {
	return m.getUserEmailByIDFn(ctx, userID)
}
func (m *mockUserProvider) EnsureUserExists(ctx context.Context, email, fullName, phone string) (uuid.UUID, error) {
	return m.ensureUserExistsFn(ctx, email, fullName, phone)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

var (
	dummyProductID = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	dummyUserID    = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	dummyAdminID   = uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	dummyGuestID   = uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
)

func defaultProductPort() *mockProductPort {
	return &mockProductPort{
		findProductByQRFn: func(_ context.Context, _ string) (uuid.UUID, error) {
			return dummyProductID, nil
		},
		validateProductOwnershipStatusFn: func(_ context.Context, _ uuid.UUID) error { return nil },
		updateOwnershipStatusFn:          func(_ context.Context, _ uuid.UUID, _ string) error { return nil },
	}
}

func defaultEmailPort(validOTP string) *mockEmailPort {
	return &mockEmailPort{
		requestOTPFn: func(_ context.Context, _, _ string) error { return nil },
		validateOTPFn: func(_ context.Context, _, otp string) (bool, error) {
			return otp == validOTP, nil
		},
	}
}

func defaultUserProvider() *mockUserProvider {
	return &mockUserProvider{
		getUserEmailByIDFn: func(_ context.Context, _ uuid.UUID) (string, string, string, error) {
			return "customer@example.com", "Nguyen Van A", "0900000001", nil
		},
		ensureUserExistsFn: func(_ context.Context, _, _, _ string) (uuid.UUID, error) {
			return dummyGuestID, nil
		},
	}
}

func defaultOwnership() *entity.Ownership {
	return &entity.Ownership{
		ID:            uuid.New(),
		ProductItemID: dummyProductID,
		OwnerID:       dummyUserID,
		Status:        entity.OwnershipStatusActive,
		OwnershipType: "PRIMARY",
		OwnedAt:       time.Now(),
	}
}

func defaultRepo() *mockOwnershipRepo {
	own := defaultOwnership()
	return &mockOwnershipRepo{
		createOwnershipFn: func(_ context.Context, _ *gorm.DB, o *entity.Ownership) (*entity.Ownership, error) {
			return o, nil
		},
		getOwnershipByProductItemIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Ownership, error) {
			return own, nil
		},
		getOwnershipHistoryByProductItemIDFn: func(_ context.Context, _ uuid.UUID) ([]entity.Ownership, error) {
			return []entity.Ownership{*own}, nil
		},
	}
}

// ---------------------------------------------------------------------------
// CUSTOMER FLOW: RequestOTP
// ---------------------------------------------------------------------------

func TestCustomer_RequestOTP_Success(t *testing.T) {
	var sentToEmail string
	emailPort := defaultEmailPort("123456")
	emailPort.requestOTPFn = func(_ context.Context, email, _ string) error {
		sentToEmail = email
		return nil
	}

	svc := NewOwnershipService(defaultRepo(), defaultProductPort(), emailPort, defaultUserProvider())
	err := svc.CustomerRequestOTP(context.Background(), dto.CustomerRequestOTPReq{QRCode: "valid-qr"}, dummyUserID)

	require.NoError(t, err)
	assert.Equal(t, "customer@example.com", sentToEmail)
}

func TestCustomer_RequestOTP_ProductNotFound(t *testing.T) {
	productPort := defaultProductPort()
	productPort.findProductByQRFn = func(_ context.Context, _ string) (uuid.UUID, error) {
		return uuid.Nil, errors.New("product not found")
	}

	svc := NewOwnershipService(defaultRepo(), productPort, defaultEmailPort("x"), defaultUserProvider())
	err := svc.CustomerRequestOTP(context.Background(), dto.CustomerRequestOTPReq{QRCode: "bad-qr"}, dummyUserID)

	require.Error(t, err)
	assert.EqualError(t, err, "product not found")
}

// ---------------------------------------------------------------------------
// CUSTOMER FLOW: VerifyAndRegister
// ---------------------------------------------------------------------------

func TestCustomer_VerifyAndRegister_Success(t *testing.T) {
	svc := NewOwnershipService(defaultRepo(), defaultProductPort(), defaultEmailPort("999999"), defaultUserProvider())

	res, err := svc.CustomerVerifyAndRegister(context.Background(), dto.CustomerVerifyAndRegisterReq{
		OTP:       "999999",
		ProductID: dummyProductID,
	}, dummyUserID)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, dummyUserID, res.OwnerID)
}

func TestCustomer_VerifyAndRegister_InvalidOTP(t *testing.T) {
	svc := NewOwnershipService(defaultRepo(), defaultProductPort(), defaultEmailPort("correct"), defaultUserProvider())

	res, err := svc.CustomerVerifyAndRegister(context.Background(), dto.CustomerVerifyAndRegisterReq{
		OTP:       "wrong",
		ProductID: dummyProductID,
	}, dummyUserID)

	require.Error(t, err)
	assert.Nil(t, res)
	assert.EqualError(t, err, "invalid or expired OTP")
}

// ---------------------------------------------------------------------------
// ADMIN FLOW: RequestOTP
// ---------------------------------------------------------------------------

func TestAdmin_RequestOTP_Success(t *testing.T) {
	var sentToEmail string
	emailPort := defaultEmailPort("123456")
	emailPort.requestOTPFn = func(_ context.Context, email, _ string) error {
		sentToEmail = email
		return nil
	}

	svc := NewOwnershipService(defaultRepo(), defaultProductPort(), emailPort, defaultUserProvider())
	err := svc.AdminRequestOTP(context.Background(), dto.AdminRequestOTPReq{
		QRCode:     "valid-qr",
		OwnerName:  "Tran Thi B",
		OwnerEmail: "tranb@example.com",
		OwnerPhone: "0911111111",
	}, dummyAdminID)

	require.NoError(t, err)
	assert.Equal(t, "tranb@example.com", sentToEmail)
}

// ---------------------------------------------------------------------------
// ADMIN FLOW: VerifyAndRegister
// ---------------------------------------------------------------------------

func TestAdmin_VerifyAndRegister_Success(t *testing.T) {
	svc := NewOwnershipService(defaultRepo(), defaultProductPort(), defaultEmailPort("888888"), defaultUserProvider())

	res, err := svc.AdminVerifyAndRegister(context.Background(), dto.AdminVerifyAndRegisterReq{
		OTP:        "888888",
		ProductID:  dummyProductID,
		OwnerName:  "Tran Thi B",
		OwnerEmail: "tranb@example.com",
		OwnerPhone: "0911111111",
	}, dummyAdminID)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, dummyGuestID, res.OwnerID)
}

func TestAdmin_VerifyAndRegister_InvalidOTP(t *testing.T) {
	svc := NewOwnershipService(defaultRepo(), defaultProductPort(), defaultEmailPort("correct"), defaultUserProvider())

	res, err := svc.AdminVerifyAndRegister(context.Background(), dto.AdminVerifyAndRegisterReq{
		OTP:        "wrong",
		ProductID:  dummyProductID,
		OwnerName:  "Tran Thi B",
		OwnerEmail: "tranb@example.com",
	}, dummyAdminID)

	require.Error(t, err)
	assert.Nil(t, res)
	assert.EqualError(t, err, "invalid or expired OTP")
}

// ---------------------------------------------------------------------------
// DETAIL API: GetOwnershipDetail (UC-P1-OWNER-02)
// ---------------------------------------------------------------------------

func TestGetOwnershipDetail_Success(t *testing.T) {
	svc := NewOwnershipService(defaultRepo(), defaultProductPort(), defaultEmailPort("x"), defaultUserProvider())

	res, err := svc.GetOwnershipDetail(context.Background(), dummyProductID)

	require.NoError(t, err)
	require.NotNil(t, res)
	// AC-001: Trả thông tin chi tiết đúng
	assert.Equal(t, dummyProductID, res.ProductID)
	assert.Equal(t, dummyUserID, res.OwnerID)
	assert.Equal(t, "Nguyen Van A", res.OwnerName)
	assert.Equal(t, "customer@example.com", res.OwnerEmail)
	assert.Equal(t, "Demo Product", res.ProductName)
	assert.Equal(t, "SKU-001", res.ProductSKU)
	// AC-002: Lịch sử phải có ít nhất 1 record
	assert.NotEmpty(t, res.OwnershipHistory)
	assert.Equal(t, "Nguyen Van A", res.OwnershipHistory[0].OwnerName)
}

func TestGetOwnershipDetail_OwnershipNotFound(t *testing.T) {
	// AC-003/VAL-002: Khi chưa có dữ liệu quyền sở hữu → trả lỗi riêng biệt
	own := defaultOwnership()
	repo := &mockOwnershipRepo{
		createOwnershipFn: func(_ context.Context, _ *gorm.DB, o *entity.Ownership) (*entity.Ownership, error) {
			return o, nil
		},
		getOwnershipByProductItemIDFn: func(_ context.Context, _ uuid.UUID) (*entity.Ownership, error) {
			return nil, gorm.ErrRecordNotFound
		},
		getOwnershipHistoryByProductItemIDFn: func(_ context.Context, _ uuid.UUID) ([]entity.Ownership, error) {
			return []entity.Ownership{*own}, nil
		},
	}
	svc := NewOwnershipService(repo, defaultProductPort(), defaultEmailPort("x"), defaultUserProvider())

	res, err := svc.GetOwnershipDetail(context.Background(), dummyProductID)

	require.Error(t, err)
	assert.Nil(t, res)
	// Thông báo lỗi theo spec: "ownership information not found"
	assert.EqualError(t, err, "ownership information not found")
}
