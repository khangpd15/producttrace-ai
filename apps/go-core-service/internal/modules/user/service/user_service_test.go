package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/entity"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/utils"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock repository
// ---------------------------------------------------------------------------

type mockUserRepository struct {
	createUserFn       func(ctx context.Context, user *entity.User) (*entity.User, error)
	getUserByEmailFn   func(ctx context.Context, email string) (*entity.User, error)
	getUserByIDFn      func(ctx context.Context, id string) (*entity.User, error)
	updateUserStatusFn func(ctx context.Context, id string, status entity.Status) error
	checkEmailExistsFn func(ctx context.Context, email string) (bool, error)
	updateUserFn       func(ctx context.Context, user *entity.User) (*entity.User, error)
	deleteUserFn       func(ctx context.Context, id string) error
	listUsersFn        func(ctx context.Context, page, limit int, role, status, search string) ([]*entity.User, int64, error)
	checkPhoneExistsFn func(ctx context.Context, phone string, excludeUserID string) (bool, error)
	writeAuditLogFn    func(ctx context.Context, content string, logType string) error
}

func (m *mockUserRepository) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	return m.createUserFn(ctx, user)
}

func (m *mockUserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return m.getUserByEmailFn(ctx, email)
}

func (m *mockUserRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	return m.getUserByIDFn(ctx, id)
}

func (m *mockUserRepository) UpdateUserStatus(ctx context.Context, id string, status entity.Status) error {
	return m.updateUserStatusFn(ctx, id, status)
}

func (m *mockUserRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	return m.checkEmailExistsFn(ctx, email)
}

func (m *mockUserRepository) UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	return m.updateUserFn(ctx, user)
}

func (m *mockUserRepository) DeleteUser(ctx context.Context, id string) error {
	return m.deleteUserFn(ctx, id)
}

func (m *mockUserRepository) ListUsers(ctx context.Context, page, limit int, role, status, search string) ([]*entity.User, int64, error) {
	return m.listUsersFn(ctx, page, limit, role, status, search)
}

func (m *mockUserRepository) CheckPhoneExists(ctx context.Context, phone string, excludeUserID string) (bool, error) {
	if m.checkPhoneExistsFn != nil {
		return m.checkPhoneExistsFn(ctx, phone, excludeUserID)
	}
	return false, nil
}

func (m *mockUserRepository) WriteAuditLog(ctx context.Context, content string, logType string) error {
	if m.writeAuditLogFn != nil {
		return m.writeAuditLogFn(ctx, content, logType)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestUserService(repo *mockUserRepository) UserServiceInterface {
	return NewUserService(repo, nil)
}

func ptrString(s string) *string {
	return &s
}

// ---------------------------------------------------------------------------
// CreateUser Tests
// ---------------------------------------------------------------------------

func TestCreateUser_Success(t *testing.T) {
	req := &request.CreateUserRequest{
		Email:    "test@example.com",
		Phone:    "0987654321",
		FullName: "Test User",
		Password: "password123",
		Role:     "CUSTOMER",
	}

	var capturedUser *entity.User
	mockRepo := &mockUserRepository{
		checkEmailExistsFn: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		createUserFn: func(ctx context.Context, user *entity.User) (*entity.User, error) {
			capturedUser = user
			return user, nil
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.CreateUser(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, req.Email, res.Email)
	assert.Equal(t, req.Phone, res.Phone)
	assert.Equal(t, req.FullName, res.FullName)
	assert.Equal(t, req.Role, res.Role)
	assert.Equal(t, string(entity.StatusActive), res.Status)
	assert.NotEmpty(t, res.ID)

	// Verify password hash
	assert.True(t, utils.ComparePassword(capturedUser.PasswordHash, req.Password))
}

func TestCreateUser_EmailExists(t *testing.T) {
	req := &request.CreateUserRequest{
		Email:    "existing@example.com",
		Phone:    "0987654321",
		FullName: "Test User",
		Password: "password123",
		Role:     "CUSTOMER",
	}

	mockRepo := &mockUserRepository{
		checkEmailExistsFn: func(ctx context.Context, email string) (bool, error) {
			return true, nil
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.CreateUser(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, res)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeConflict, appErr.Code)
}

func TestCreateUser_CheckEmailError(t *testing.T) {
	req := &request.CreateUserRequest{
		Email:    "test@example.com",
		Phone:    "0987654321",
		FullName: "Test User",
		Password: "password123",
		Role:     "CUSTOMER",
	}

	dbErr := errors.New("db connection failure")
	mockRepo := &mockUserRepository{
		checkEmailExistsFn: func(ctx context.Context, email string) (bool, error) {
			return false, dbErr
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.CreateUser(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, dbErr, err)
}

func TestCreateUser_RepoCreateError(t *testing.T) {
	req := &request.CreateUserRequest{
		Email:    "test@example.com",
		Phone:    "0987654321",
		FullName: "Test User",
		Password: "password123",
		Role:     "CUSTOMER",
	}

	dbErr := errors.New("db save failure")
	mockRepo := &mockUserRepository{
		checkEmailExistsFn: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		createUserFn: func(ctx context.Context, user *entity.User) (*entity.User, error) {
			return nil, dbErr
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.CreateUser(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, dbErr, err)
}

// ---------------------------------------------------------------------------
// GetUserByID Tests
// ---------------------------------------------------------------------------

func TestGetUserByID_Success(t *testing.T) {
	userID := uuid.New()
	existingUser := &entity.User{
		ID:        userID,
		Email:     "test@example.com",
		Phone:     "0987654321",
		FullName:  "Test User",
		Role:      entity.RoleCustomer,
		Status:    entity.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			assert.Equal(t, userID.String(), id)
			return existingUser, nil
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.GetUserByID(context.Background(), userID.String())

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, userID.String(), res.ID)
	assert.Equal(t, existingUser.Email, res.Email)
}

func TestGetUserByID_NotFound(t *testing.T) {
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return nil, nil
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.GetUserByID(context.Background(), uuid.New().String())

	require.Error(t, err)
	assert.Nil(t, res)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeNotFound, appErr.Code)
}

func TestGetUserByID_RepoError(t *testing.T) {
	dbErr := errors.New("db fetch failure")
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return nil, dbErr
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.GetUserByID(context.Background(), uuid.New().String())

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, dbErr, err)
}

// ---------------------------------------------------------------------------
// UpdateUser Tests
// ---------------------------------------------------------------------------

func TestUpdateUser_Success_NoPassword(t *testing.T) {
	userID := uuid.New()
	existingUser := &entity.User{
		ID:           userID,
		Email:        "old@example.com",
		Phone:        "0987654321",
		FullName:     "Old Name",
		PasswordHash: "some_old_hash",
		Role:         entity.RoleCustomer,
		Status:       entity.StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	req := &request.UpdateUserRequest{
		Email:    "new@example.com",
		Phone:    "0123456789",
		FullName: "New Name",
		Role:     "ADMIN",
		Status:   "BANNED",
		Password: "", // Do not update password
	}

	var capturedUser *entity.User
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return existingUser, nil
		},
		checkEmailExistsFn: func(ctx context.Context, email string) (bool, error) {
			assert.Equal(t, req.Email, email)
			return false, nil
		},
		updateUserFn: func(ctx context.Context, user *entity.User) (*entity.User, error) {
			capturedUser = user
			return user, nil
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.UpdateUser(context.Background(), userID.String(), req)

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, req.Email, res.Email)
	assert.Equal(t, req.Phone, res.Phone)
	assert.Equal(t, req.FullName, res.FullName)
	assert.Equal(t, req.Role, res.Role)
	assert.Equal(t, req.Status, res.Status)

	// Password hash should not be changed
	assert.Equal(t, "some_old_hash", capturedUser.PasswordHash)
}

func TestUpdateUser_Success_WithPassword(t *testing.T) {
	userID := uuid.New()
	existingUser := &entity.User{
		ID:           userID,
		Email:        "old@example.com",
		Phone:        "0987654321",
		FullName:     "Old Name",
		PasswordHash: "some_old_hash",
		Role:         entity.RoleCustomer,
		Status:       entity.StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	req := &request.UpdateUserRequest{
		Email:    "old@example.com", // Email not changed, so CheckEmailExists should not be called
		Phone:    "0123456789",
		FullName: "New Name",
		Role:     "ADMIN",
		Status:   "BANNED",
		Password: "newsecurepassword",
	}

	var capturedUser *entity.User
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return existingUser, nil
		},
		checkEmailExistsFn: func(ctx context.Context, email string) (bool, error) {
			t.Fatal("CheckEmailExists should not be called because email has not changed")
			return false, nil
		},
		updateUserFn: func(ctx context.Context, user *entity.User) (*entity.User, error) {
			capturedUser = user
			return user, nil
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.UpdateUser(context.Background(), userID.String(), req)

	require.NoError(t, err)
	require.NotNil(t, res)

	// Password hash should be updated and match the new password
	assert.NotEqual(t, "some_old_hash", capturedUser.PasswordHash)
	assert.True(t, utils.ComparePassword(capturedUser.PasswordHash, req.Password))
}

func TestUpdateUser_NotFound(t *testing.T) {
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return nil, nil
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.UpdateUser(context.Background(), uuid.New().String(), &request.UpdateUserRequest{})

	require.Error(t, err)
	assert.Nil(t, res)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeNotFound, appErr.Code)
}

func TestUpdateUser_EmailConflict(t *testing.T) {
	userID := uuid.New()
	existingUser := &entity.User{
		ID:        userID,
		Email:     "old@example.com",
		Role:      entity.RoleCustomer,
		Status:    entity.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	req := &request.UpdateUserRequest{
		Email:    "conflict@example.com",
		Role:     "CUSTOMER",
		Status:   "ACTIVE",
		Password: "",
	}

	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return existingUser, nil
		},
		checkEmailExistsFn: func(ctx context.Context, email string) (bool, error) {
			return true, nil
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.UpdateUser(context.Background(), userID.String(), req)

	require.Error(t, err)
	assert.Nil(t, res)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeConflict, appErr.Code)
}

func TestUpdateUser_GetUserByIDError(t *testing.T) {
	dbErr := errors.New("db error")
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return nil, dbErr
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.UpdateUser(context.Background(), uuid.New().String(), &request.UpdateUserRequest{})

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, dbErr, err)
}

func TestUpdateUser_CheckEmailError(t *testing.T) {
	userID := uuid.New()
	existingUser := &entity.User{
		ID:    userID,
		Email: "old@example.com",
	}

	req := &request.UpdateUserRequest{
		Email: "new@example.com",
	}

	dbErr := errors.New("db error")
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return existingUser, nil
		},
		checkEmailExistsFn: func(ctx context.Context, email string) (bool, error) {
			return false, dbErr
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.UpdateUser(context.Background(), userID.String(), req)

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, dbErr, err)
}

func TestUpdateUser_UpdateRepoError(t *testing.T) {
	userID := uuid.New()
	existingUser := &entity.User{
		ID:    userID,
		Email: "old@example.com",
	}

	req := &request.UpdateUserRequest{
		Email: "old@example.com",
	}

	dbErr := errors.New("db update error")
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return existingUser, nil
		},
		updateUserFn: func(ctx context.Context, user *entity.User) (*entity.User, error) {
			return nil, dbErr
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.UpdateUser(context.Background(), userID.String(), req)

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, dbErr, err)
}

// ---------------------------------------------------------------------------
// DeleteUser Tests
// ---------------------------------------------------------------------------

func TestDeleteUser_Success(t *testing.T) {
	userID := uuid.New()
	existingUser := &entity.User{
		ID: userID,
	}

	var deleteCalledWith string
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return existingUser, nil
		},
		deleteUserFn: func(ctx context.Context, id string) error {
			deleteCalledWith = id
			return nil
		},
	}

	svc := newTestUserService(mockRepo)
	err := svc.DeleteUser(context.Background(), userID.String())

	require.NoError(t, err)
	assert.Equal(t, userID.String(), deleteCalledWith)
}

func TestDeleteUser_NotFound(t *testing.T) {
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return nil, nil
		},
	}

	svc := newTestUserService(mockRepo)
	err := svc.DeleteUser(context.Background(), uuid.New().String())

	require.Error(t, err)
	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeNotFound, appErr.Code)
}

func TestDeleteUser_GetUserError(t *testing.T) {
	dbErr := errors.New("db get error")
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return nil, dbErr
		},
	}

	svc := newTestUserService(mockRepo)
	err := svc.DeleteUser(context.Background(), uuid.New().String())

	require.Error(t, err)
	assert.Equal(t, dbErr, err)
}

func TestDeleteUser_DeleteRepoError(t *testing.T) {
	userID := uuid.New()
	existingUser := &entity.User{
		ID: userID,
	}

	dbErr := errors.New("db delete error")
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return existingUser, nil
		},
		deleteUserFn: func(ctx context.Context, id string) error {
			return dbErr
		},
	}

	svc := newTestUserService(mockRepo)
	err := svc.DeleteUser(context.Background(), userID.String())

	require.Error(t, err)
	assert.Equal(t, dbErr, err)
}

// ---------------------------------------------------------------------------
// ListUsers Tests
// ---------------------------------------------------------------------------

func TestListUsers_Success(t *testing.T) {
	user1 := &entity.User{ID: uuid.New(), Email: "u1@example.com"}
	user2 := &entity.User{ID: uuid.New(), Email: "u2@example.com"}
	usersList := []*entity.User{user1, user2}

	var capturedPage, capturedLimit int
	var capturedRole, capturedStatus, capturedSearch string

	mockRepo := &mockUserRepository{
		listUsersFn: func(ctx context.Context, page, limit int, role, status, search string) ([]*entity.User, int64, error) {
			capturedPage = page
			capturedLimit = limit
			capturedRole = role
			capturedStatus = status
			capturedSearch = search
			return usersList, 42, nil
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.ListUsers(context.Background(), 2, 5, "ADMIN", "ACTIVE", "John")

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, 2, res.Page)
	assert.Equal(t, 5, res.Limit)
	assert.Equal(t, int64(42), res.Total)
	require.Len(t, res.Items, 2)
	assert.Equal(t, user1.ID.String(), res.Items[0].ID)
	assert.Equal(t, user2.ID.String(), res.Items[1].ID)

	assert.Equal(t, 2, capturedPage)
	assert.Equal(t, 5, capturedLimit)
	assert.Equal(t, "ADMIN", capturedRole)
	assert.Equal(t, "ACTIVE", capturedStatus)
	assert.Equal(t, "John", capturedSearch)
}

func TestListUsers_Defaults(t *testing.T) {
	mockRepo := &mockUserRepository{
		listUsersFn: func(ctx context.Context, page, limit int, role, status, search string) ([]*entity.User, int64, error) {
			assert.Equal(t, 1, page)
			assert.Equal(t, 10, limit)
			return []*entity.User{}, 0, nil
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.ListUsers(context.Background(), 0, -5, "", "", "")

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 1, res.Page)
	assert.Equal(t, 10, res.Limit)
}

func TestListUsers_RepoError(t *testing.T) {
	dbErr := errors.New("db error")
	mockRepo := &mockUserRepository{
		listUsersFn: func(ctx context.Context, page, limit int, role, status, search string) ([]*entity.User, int64, error) {
			return nil, 0, dbErr
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.ListUsers(context.Background(), 1, 10, "", "", "")

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, dbErr, err)
}

// ---------------------------------------------------------------------------
// UpdateProfile Tests
// ---------------------------------------------------------------------------

func TestUpdateProfile_Success(t *testing.T) {
	userID := uuid.New()
	existingUser := &entity.User{
		ID:        userID,
		Email:     "test@example.com",
		Phone:     "0000000000",
		FullName:  "Old Name",
		Role:      entity.RoleCustomer,
		Status:    entity.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	req := &request.UpdateProfileRequest{
		FullName: ptrString("New Name"),
		Phone:    ptrString("0111111111"),
	}

	var capturedUser *entity.User
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return existingUser, nil
		},
		checkPhoneExistsFn: func(ctx context.Context, phone string, excludeUserID string) (bool, error) {
			return false, nil
		},
		writeAuditLogFn: func(ctx context.Context, content string, logType string) error {
			return nil
		},
		updateUserFn: func(ctx context.Context, user *entity.User) (*entity.User, error) {
			capturedUser = user
			return user, nil
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.UpdateProfile(context.Background(), userID.String(), userID.String(), req)

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, *req.FullName, res.FullName)
	assert.Equal(t, *req.Phone, res.Phone)
	assert.Equal(t, existingUser.Email, res.Email)

	assert.Equal(t, *req.FullName, capturedUser.FullName)
	assert.Equal(t, *req.Phone, capturedUser.Phone)
}

func TestUpdateProfile_NotFound(t *testing.T) {
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return nil, nil
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.UpdateProfile(context.Background(), uuid.New().String(), uuid.New().String(), &request.UpdateProfileRequest{})

	require.Error(t, err)
	assert.Nil(t, res)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeNotFound, appErr.Code)
}

func TestUpdateProfile_GetUserError(t *testing.T) {
	dbErr := errors.New("db error")
	mockRepo := &mockUserRepository{
		getUserByIDFn: func(ctx context.Context, id string) (*entity.User, error) {
			return nil, dbErr
		},
	}

	svc := newTestUserService(mockRepo)
	res, err := svc.UpdateProfile(context.Background(), uuid.New().String(), uuid.New().String(), &request.UpdateProfileRequest{})

	require.Error(t, err)
	assert.Nil(t, res)
}

// ---------------------------------------------------------------------------
// SearchUsers Tests (FR-011)
// ---------------------------------------------------------------------------

func TestSearchUsers_Success_WithAllFilters(t *testing.T) {
	user1 := &entity.User{ID: uuid.New(), Email: "john@example.com", FullName: "John Doe", Role: entity.RoleCustomer, Status: entity.StatusActive}
	user2 := &entity.User{ID: uuid.New(), Email: "john.smith@example.com", FullName: "John Smith", Role: entity.RoleCustomer, Status: entity.StatusActive}
	usersList := []*entity.User{user1, user2}

	var capturedPage, capturedLimit int
	var capturedRole, capturedStatus, capturedSearch string

	mockRepo := &mockUserRepository{
		listUsersFn: func(ctx context.Context, page, limit int, role, status, search string) ([]*entity.User, int64, error) {
			capturedPage = page
			capturedLimit = limit
			capturedRole = role
			capturedStatus = status
			capturedSearch = search
			return usersList, 2, nil
		},
	}

	svc := newTestUserService(mockRepo)
	req := &request.SearchUserRequest{
		Keyword: "John",
		Role:    "CUSTOMER",
		Status:  "ACTIVE",
		Page:    1,
		Limit:   10,
	}
	res, err := svc.SearchUsers(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, 1, res.Page)
	assert.Equal(t, 10, res.Limit)
	assert.Equal(t, int64(2), res.Total)
	require.Len(t, res.Items, 2)
	assert.Equal(t, user1.ID.String(), res.Items[0].ID)
	assert.Equal(t, user2.ID.String(), res.Items[1].ID)

	assert.Equal(t, 1, capturedPage)
	assert.Equal(t, 10, capturedLimit)
	assert.Equal(t, "CUSTOMER", capturedRole)
	assert.Equal(t, "ACTIVE", capturedStatus)
	assert.Equal(t, "John", capturedSearch)
}

func TestSearchUsers_Success_KeywordOnly(t *testing.T) {
	user1 := &entity.User{ID: uuid.New(), Email: "test@example.com", FullName: "Test User"}
	usersList := []*entity.User{user1}

	mockRepo := &mockUserRepository{
		listUsersFn: func(ctx context.Context, page, limit int, role, status, search string) ([]*entity.User, int64, error) {
			assert.Equal(t, "test", search)
			assert.Equal(t, "", role)
			assert.Equal(t, "", status)
			return usersList, 1, nil
		},
	}

	svc := newTestUserService(mockRepo)
	req := &request.SearchUserRequest{
		Keyword: "test",
	}
	res, err := svc.SearchUsers(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Items, 1)
	assert.Equal(t, user1.ID.String(), res.Items[0].ID)
}

func TestSearchUsers_Success_EmptyResults(t *testing.T) {
	mockRepo := &mockUserRepository{
		listUsersFn: func(ctx context.Context, page, limit int, role, status, search string) ([]*entity.User, int64, error) {
			return []*entity.User{}, 0, nil
		},
	}

	svc := newTestUserService(mockRepo)
	req := &request.SearchUserRequest{
		Keyword: "nonexistent",
	}
	res, err := svc.SearchUsers(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, int64(0), res.Total)
	assert.Empty(t, res.Items)
}

func TestSearchUsers_KeywordTooLong(t *testing.T) {
	longKeyword := strings.Repeat("a", 256)

	mockRepo := &mockUserRepository{}

	svc := newTestUserService(mockRepo)
	req := &request.SearchUserRequest{
		Keyword: longKeyword,
	}
	res, err := svc.SearchUsers(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, res)

	var appErr *apperror.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperror.CodeValidation, appErr.Code)
}

func TestSearchUsers_DefaultPagination(t *testing.T) {
	mockRepo := &mockUserRepository{
		listUsersFn: func(ctx context.Context, page, limit int, role, status, search string) ([]*entity.User, int64, error) {
			assert.Equal(t, 1, page)
			assert.Equal(t, 10, limit)
			return []*entity.User{}, 0, nil
		},
	}

	svc := newTestUserService(mockRepo)
	req := &request.SearchUserRequest{
		Keyword: "test",
		Page:    0,
		Limit:   -5,
	}
	res, err := svc.SearchUsers(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 1, res.Page)
	assert.Equal(t, 10, res.Limit)
}

func TestSearchUsers_RepoError(t *testing.T) {
	dbErr := errors.New("db connection failure")
	mockRepo := &mockUserRepository{
		listUsersFn: func(ctx context.Context, page, limit int, role, status, search string) ([]*entity.User, int64, error) {
			return nil, 0, dbErr
		},
	}

	svc := newTestUserService(mockRepo)
	req := &request.SearchUserRequest{
		Keyword: "test",
		Page:    1,
		Limit:   10,
	}
	res, err := svc.SearchUsers(context.Background(), req)

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, dbErr, err)
}

func TestSearchUsers_ByRoleOnly(t *testing.T) {
	user1 := &entity.User{ID: uuid.New(), Email: "admin@example.com", FullName: "Admin User", Role: entity.RoleAdmin, Status: entity.StatusActive}
	usersList := []*entity.User{user1}

	mockRepo := &mockUserRepository{
		listUsersFn: func(ctx context.Context, page, limit int, role, status, search string) ([]*entity.User, int64, error) {
			assert.Equal(t, "ADMIN", role)
			assert.Equal(t, "", status)
			assert.Equal(t, "", search)
			return usersList, 1, nil
		},
	}

	svc := newTestUserService(mockRepo)
	req := &request.SearchUserRequest{
		Role: "ADMIN",
	}
	res, err := svc.SearchUsers(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Items, 1)
	assert.Equal(t, "ADMIN", res.Items[0].Role)
}

func TestSearchUsers_ByStatusOnly(t *testing.T) {
	user1 := &entity.User{ID: uuid.New(), Email: "active@example.com", FullName: "Active User", Role: entity.RoleCustomer, Status: entity.StatusActive}
	usersList := []*entity.User{user1}

	mockRepo := &mockUserRepository{
		listUsersFn: func(ctx context.Context, page, limit int, role, status, search string) ([]*entity.User, int64, error) {
			assert.Equal(t, "", role)
			assert.Equal(t, "ACTIVE", status)
			assert.Equal(t, "", search)
			return usersList, 1, nil
		},
	}

	svc := newTestUserService(mockRepo)
	req := &request.SearchUserRequest{
		Status: "ACTIVE",
	}
	res, err := svc.SearchUsers(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Items, 1)
	assert.Equal(t, "ACTIVE", res.Items[0].Status)
}

func TestSearchUsers_EmptyRequest(t *testing.T) {
	user1 := &entity.User{ID: uuid.New(), Email: "u1@example.com"}
	user2 := &entity.User{ID: uuid.New(), Email: "u2@example.com"}
	usersList := []*entity.User{user1, user2}

	mockRepo := &mockUserRepository{
		listUsersFn: func(ctx context.Context, page, limit int, role, status, search string) ([]*entity.User, int64, error) {
			assert.Equal(t, "", role)
			assert.Equal(t, "", status)
			assert.Equal(t, "", search)
			assert.Equal(t, 1, page)
			assert.Equal(t, 10, limit)
			return usersList, 2, nil
		},
	}

	svc := newTestUserService(mockRepo)
	req := &request.SearchUserRequest{}
	res, err := svc.SearchUsers(context.Background(), req)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, int64(2), res.Total)
	require.Len(t, res.Items, 2)
}
