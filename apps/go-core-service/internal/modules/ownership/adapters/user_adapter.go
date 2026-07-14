package adapters

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/service"
	UserEntity "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/entity"
	userRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/repository"
)

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
