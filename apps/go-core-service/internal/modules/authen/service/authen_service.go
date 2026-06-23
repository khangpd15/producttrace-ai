package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/publisher"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/rabbitmq"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/types"
	UserEntity "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/entity"
	UserRepository "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/repository"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/utils"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/cache"
)

type AuthenServiceInterface interface {
	RegisterUser(ctx context.Context, email, phone, fullName, password string) (*UserEntity.User, error)
	LoginUser(ctx context.Context, email, password string) (accessToken string, refreshToken string, err error)
	VerifyOTP(ctx context.Context, email, otp string) error
	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, err error)
}

type AuthenService struct {
	userRepository UserRepository.UserRepositoryInterface
	cache          cache.Cache
	publisher      *publisher.Publisher
}

func NewAuthenService(
	userRepository UserRepository.UserRepositoryInterface,
	cache cache.Cache,
	publisher *publisher.Publisher,
) AuthenServiceInterface {
	return &AuthenService{
		userRepository: userRepository,
		cache:          cache,
		publisher:      publisher,
	}
}

func (s *AuthenService) RegisterUser(ctx context.Context, email, phone, fullName, password string) (*UserEntity.User, error) {
	// Check if user already exists
	exists, err := s.userRepository.CheckEmailExists(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.NewConflict("User already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create new user
	newUser := UserEntity.NewUser(email, phone, fullName, hashedPassword, string(UserEntity.RoleCustomer))
	savedUser, err := s.userRepository.CreateUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	// Generate OTP
	otpCode, err := utils.GenerateOTP()
	if err != nil {
		return nil, err
	}

	// Save OTP to Redis with 5m TTL
	err = s.cache.Set(ctx, fmt.Sprintf("otp:email:%s", email), otpCode, 5*time.Minute)
	if err != nil {
		return nil, err
	}

	// Publish event to RabbitMQ
	event := types.Event{
		EventID:       uuid.New().String(),
		EventType:     rabbitmq.UserRegisteredRK,
		EventVersion:  "1.0.0",
		Timestamp:     time.Now().UTC(),
		Producer:      "go-core-service",
		CorrelationID: uuid.New().String(),
		Payload: map[string]interface{}{
			"email":     savedUser.Email,
			"full_name": savedUser.FullName,
			"otp_code":  otpCode,
		},
	}

	// Publish user.registered event
	_ = s.publisher.Publish(event)

	return savedUser, nil
}

func (s *AuthenService) LoginUser(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", apperror.NewUnauthorized("Invalid credentials")
	}

	// Check if status is pending verification
	if user.Status == UserEntity.StatusPending {
		return "", "", apperror.NewValidation("Please verify your account via OTP first")
	}

	if user.Status != UserEntity.StatusActive {
		return "", "", apperror.NewUnauthorized("Account is not active")
	}

	// Compare passwords
	if !utils.ComparePassword(user.PasswordHash, password) {
		return "", "", apperror.NewUnauthorized("Invalid credentials")
	}

	// Generate JWT Access Token
	accessToken, err := utils.GenerateAccessToken(user.ID.String(), user.Email, string(user.Role))
	if err != nil {
		return "", "", err
	}

	// Generate Refresh Token
	rawToken := uuid.NewString()
	hashedToken := utils.HashToken(rawToken)

	key := fmt.Sprintf(
		"refresh_token:%s:%s",
		user.ID.String(),
		hashedToken,
	)

	err = s.cache.Set(
		ctx,
		key,
		user.ID.String(),
		7*24*time.Hour,
	)
	if err != nil {
		return "", "", err
	}

	clientRefreshToken := fmt.Sprintf("%s.%s", user.ID.String(), rawToken)
	return accessToken, clientRefreshToken, nil
}

func (s *AuthenService) VerifyOTP(ctx context.Context, email, otp string) error {
	otpKey := fmt.Sprintf("otp:email:%s", email)
	storedOtp, err := s.cache.Get(ctx, otpKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return apperror.NewValidation("OTP has expired or is invalid")
		}
		return err
	}

	if storedOtp != otp {
		return apperror.NewValidation("Incorrect OTP code")
	}

	// OTP matches, delete it
	s.cache.Delete(ctx, otpKey)

	// Fetch user
	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.NewNotFound("User not found")
	}

	// Update user status to ACTIVE
	err = s.userRepository.UpdateUserStatus(ctx, user.ID.String(), UserEntity.StatusActive)
	if err != nil {
		return err
	}

	// Publish user.verified event
	event := types.Event{
		EventID:       uuid.New().String(),
		EventType:     rabbitmq.UserVerifiedRK,
		EventVersion:  "1.0.0",
		Timestamp:     time.Now().UTC(),
		Producer:      "go-core-service",
		CorrelationID: uuid.New().String(),
		Payload: map[string]interface{}{
			"email": user.Email,
		},
	}
	_ = s.publisher.Publish(event)

	return nil
}

func (s *AuthenService) RefreshToken(ctx context.Context, oldRefreshToken string) (string, string, error) {
	parts := strings.Split(oldRefreshToken, ".")
	if len(parts) != 2 {
		return "", "", apperror.NewUnauthorized("Invalid refresh token format")
	}
	userID := parts[0]
	rawToken := parts[1]

	hashedToken := utils.HashToken(rawToken)
	tokenKey := fmt.Sprintf("refresh_token:%s:%s", userID, hashedToken)

	storedUserID, err := s.cache.Get(ctx, tokenKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", "", apperror.NewUnauthorized("Invalid or expired refresh token")
		}
		return "", "", err
	}

	// Delete old refresh token (Token Rotation)
	s.cache.Delete(ctx, tokenKey)

	// Fetch user to verify active status
	user, err := s.userRepository.GetUserByID(ctx, storedUserID)
	if err != nil {
		return "", "", err
	}
	if user == nil || user.Status != UserEntity.StatusActive {
		return "", "", apperror.NewUnauthorized("User is inactive or not found")
	}

	// Generate new Access Token
	newAccessToken, err := utils.GenerateAccessToken(user.ID.String(), user.Email, string(user.Role))
	if err != nil {
		return "", "", err
	}

	// Generate new Refresh Token
	newRawToken := uuid.NewString()
	newHashedToken := utils.HashToken(newRawToken)
	newTokenKey := fmt.Sprintf("refresh_token:%s:%s", user.ID.String(), newHashedToken)

	// Store new Refresh Token in Redis
	err = s.cache.Set(ctx, newTokenKey, user.ID.String(), 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	newClientRefreshToken := fmt.Sprintf("%s.%s", user.ID.String(), newRawToken)
	return newAccessToken, newClientRefreshToken, nil
}
