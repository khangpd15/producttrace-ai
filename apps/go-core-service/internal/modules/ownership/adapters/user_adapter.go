package adapters

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/service"
)

type DummyUserAdapter struct{}

func NewDummyUserAdapter() service.IUserInfoProvider {
	return &DummyUserAdapter{}
}

func (a *DummyUserAdapter) GetUserEmailByID(ctx context.Context, userID uuid.UUID) (email string, fullName string, phone string, err error) {
	// Dummy: returns fake profile data until User Repository integration is done
	email = fmt.Sprintf("customer-%s@example.com", userID.String()[:8])
	fullName = "Khách Hàng Demo"
	phone = "0900000000"
	return
}

func (a *DummyUserAdapter) EnsureUserExists(ctx context.Context, email, fullName, phone string) (uuid.UUID, error) {
	// Dummy: returning a mock user ID for Admin flow
	return uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd"), nil
}

func (a *DummyUserAdapter) SearchUserIDs(ctx context.Context, name string, email string, phone string) ([]uuid.UUID, error) {
	log.Printf("[UserAdapter] Faking SearchUserIDs for Name: %s, Email: %s\n", name, email)
	return []uuid.UUID{uuid.New(), uuid.New()}, nil
}
