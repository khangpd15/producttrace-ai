package adapters

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/service"
	UserEntity "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/entity"
	userRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/repository"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
)

type RealUserAdapter struct {
	userRepo userRepo.UserRepositoryInterface
}

func NewRealUserAdapter(userRepo userRepo.UserRepositoryInterface) service.IUserInfoProvider {
	return &RealUserAdapter{userRepo: userRepo}
}

func (a *RealUserAdapter) GetUserEmailByID(ctx context.Context, userID uuid.UUID) (email string, fullName string, phone string, err error) {
	user, err := a.userRepo.GetUserByID(ctx, userID.String())
	if err != nil {
		return "", "", "", err
	}
	if user == nil {
		return "", "", "", apperror.NewNotFound("User")
	}
	return user.Email, user.FullName, user.Phone, nil
}

func (a *RealUserAdapter) EnsureUserExists(ctx context.Context, email, fullName, phone string) (uuid.UUID, error) {
	existing, err := a.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return uuid.Nil, err
	}
	if existing == nil {
		return uuid.Nil, apperror.NewBadRequest("Tài khoản email này chưa được đăng ký trong hệ thống. Bạn/khách hàng cần đăng ký tài khoản trước!")
	}
	return existing.ID, nil
}

func (a *RealUserAdapter) GetUserByEmail(ctx context.Context, email string) (uuid.UUID, string, error) {
	user, err := a.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return uuid.Nil, "", err
	}
	if user == nil {
		return uuid.Nil, "", apperror.NewNotFound("User")
	}
	return user.ID, string(user.Status), nil
}

func (a *RealUserAdapter) SearchUserIDs(ctx context.Context, name string, email string, phone string) ([]uuid.UUID, error) {
	users, err := a.userRepo.SearchUsers(ctx, name, email, phone)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
	}

	return ids, nil
}

// -------------------------------------------------------------
// EXISTING DummyUserAdapter FROM DEVELOP
// -------------------------------------------------------------

type DummyUserAdapter struct {
	uRepo userRepo.UserRepositoryInterface
}

func NewDummyUserAdapter(uRepo userRepo.UserRepositoryInterface) service.IUserInfoProvider {
	return &DummyUserAdapter{uRepo: uRepo}
}

func (a *DummyUserAdapter) GetUserEmailByID(ctx context.Context, userID uuid.UUID) (email string, fullName string, phone string, err error) {
	user, err := a.uRepo.GetUserByID(ctx, userID.String())
	if err != nil {
		return "", "", "", err
	}
	if user == nil {
		return "", "", "", fmt.Errorf("user not found")
	}
	return user.Email, user.FullName, user.Phone, nil
}

func (a *DummyUserAdapter) EnsureUserExists(ctx context.Context, email, fullName, phone string) (uuid.UUID, error) {
	user, err := a.uRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return uuid.Nil, err
	}
	if user != nil {
		return user.ID, nil
	}

	// Create user if not exists
	newUser := UserEntity.NewUser(email, phone, fullName, "", "CUSTOMER")
	newUser.Status = UserEntity.StatusActive // Mark as active for guest
	created, err := a.uRepo.CreateUser(ctx, newUser)
	if err != nil {
		return uuid.Nil, err
	}
	return created.ID, nil
}

func (a *DummyUserAdapter) GetUserByEmail(ctx context.Context, email string) (uuid.UUID, string, error) {
	user, err := a.uRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return uuid.Nil, "", err
	}
	if user == nil {
		return uuid.Nil, "", fmt.Errorf("user not found")
	}
	return user.ID, string(user.Status), nil
}

func (a *DummyUserAdapter) SearchUserIDs(ctx context.Context, name string, email string, phone string) ([]uuid.UUID, error) {
	searchQuery := name
	if searchQuery == "" {
		searchQuery = email
	}
	if searchQuery == "" {
		searchQuery = phone
	}
	users, _, err := a.uRepo.ListUsers(ctx, 1, 100, "", "", searchQuery)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	return ids, nil
}
