package service

import (
	"context"

	"github.com/google/uuid"
)

// IUserInfoProvider is used to fetch user profile info when a Customer registers.
// This avoids Customer needing to fill in their own name/email/phone again.
type IUserInfoProvider interface {
	GetUserEmailByID(ctx context.Context, userID uuid.UUID) (email string, fullName string, phone string, err error)

	// EnsureUserExists is used by Admin flow: searches for user by email.
	// If not found, it creates a placeholder/guest user account so we can link ownership.
	EnsureUserExists(ctx context.Context, email, fullName, phone string) (ownerID uuid.UUID, err error)

	GetUserByEmail(ctx context.Context, email string) (userID uuid.UUID, status string, err error)

	// SearchUserIDs dùng để filter search ownership (FR-042)
	SearchUserIDs(ctx context.Context, name string, email string, phone string) ([]uuid.UUID, error)
}
