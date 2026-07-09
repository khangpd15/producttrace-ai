package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/consumer"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/publisher"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/rabbitmq"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/types"
	UserEntity "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/entity"
	UserRepository "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/repository"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/utils"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/cache"
)

const (
	asyncPublishTimeout = 4 * time.Second
)

type AuthenServiceInterface interface {
	RegisterUser(ctx context.Context, email, phone, fullName, password string) (*UserEntity.User, error)
	LoginUser(ctx context.Context, email, password string) (accessToken string, refreshToken string, err error)
	VerifyOTP(ctx context.Context, email, otp string) error
	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, err error)
	Logout(ctx context.Context, refreshToken string) error
	ReSendOTP(ctx context.Context, email string) error
	ConsumerOTPEvent(ctx context.Context) error
}

type AuthenService struct {
	userRepository UserRepository.UserRepositoryInterface
	cache          cache.Cache
	publisher      *publisher.Publisher
	consumer       *consumer.Consumer
}

type otpEventEnvelope struct {
	Pattern string `json:"pattern"`
	Data    struct {
		EventType     string          `json:"event_type"`
		EventVersion  string          `json:"event_version"`
		CorrelationID string          `json:"correlation_id"`
		Payload       json.RawMessage `json:"payload"`
	} `json:"data"`
}

type userRegisteredPayload struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone,omitempty"`
	OTPCode  string `json:"otp_code,omitempty"`
}

func NewAuthenService(
	userRepository UserRepository.UserRepositoryInterface,
	cache cache.Cache,
	publisher *publisher.Publisher,
	consumers ...*consumer.Consumer,
) AuthenServiceInterface {
	var authConsumer *consumer.Consumer
	if len(consumers) > 0 {
		authConsumer = consumers[0]
	}

	return &AuthenService{
		userRepository: userRepository,
		cache:          cache,
		publisher:      publisher,
		consumer:       authConsumer,
	}
}

func (s *AuthenService) RegisterUser(ctx context.Context, email, phone, fullName, password string) (*UserEntity.User, error) {

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := UserEntity.NewUser(email, phone, fullName, hashedPassword, string(UserEntity.RoleCustomer))
	savedUser, err := s.userRepository.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	correlationID := uuid.New().String()

	event := types.Event{
		EventID:       uuid.New().String(),
		EventType:     rabbitmq.UserRegisteredRK,
		EventVersion:  "1.0.0",
		Timestamp:     time.Now().UTC(),
		Producer:      "go-core-service",
		CorrelationID: correlationID,
		Payload: map[string]interface{}{
			"email":     savedUser.Email,
			"full_name": savedUser.FullName,
			"phone":     savedUser.Phone,
		},
	}

	if err := s.publisher.PublishWithContext(ctx, event); err != nil {
		return nil, apperror.NewInternal("publish user.registered failed")
	}

	log.Printf("[REGISTER] user created email=%s", email)

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

	if user.Status == UserEntity.StatusPending {
		return "", "", apperror.NewValidation("Please verify your account via OTP first")
	}

	if user.Status != UserEntity.StatusActive {
		return "", "", apperror.NewUnauthorized("Account is not active")
	}

	if !utils.ComparePassword(user.PasswordHash, password) {
		return "", "", apperror.NewUnauthorized("Invalid credentials")
	}

	accessToken, err := utils.GenerateAccessToken(user.ID.String(), user.Email, string(user.Role))
	if err != nil {
		return "", "", err
	}

	rawToken := uuid.NewString()
	hashedToken := utils.HashToken(rawToken)

	key := fmt.Sprintf("refresh_token:%s:%s", user.ID.String(), hashedToken)
	err = s.cache.Set(ctx, key, user.ID.String(), 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	clientRefreshToken := fmt.Sprintf("%s.%s", user.ID.String(), rawToken)
	return accessToken, clientRefreshToken, nil
}

func (s *AuthenService) ReSendOTP(ctx context.Context, email string) error {

	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.NewNotFound("User not found")
	}

	otpCode, err := utils.GenerateOTP()
	if err != nil {
		return err
	}

	if err := s.cache.Set(ctx, fmt.Sprintf("otp:email:%s", email), otpCode, 5*time.Minute); err != nil {
		return err
	}

	event := types.Event{
		EventID:       uuid.New().String(),
		EventType:     rabbitmq.OTPRegisterUserRK,
		EventVersion:  "1.0.0",
		Timestamp:     time.Now().UTC(),
		Producer:      "go-core-service",
		CorrelationID: uuid.New().String(),
		Payload: map[string]interface{}{
			"email":     user.Email,
			"full_name": user.FullName,
			"otp_code":  otpCode,
		},
	}

	if err := s.publisher.PublishWithContext(ctx, event); err != nil {
		return err
	}

	return nil
}

func (s *AuthenService) VerifyOTP(ctx context.Context, email, otp string) error {
	otpKey := fmt.Sprintf("otp:email:%s", email)
	storedOtp, err := s.cache.Get(ctx, otpKey)
	if err != nil {
		if errors.Is(err, redis.Nil) || err.Error() == "redis: nil" {
			return apperror.NewValidation("OTP has expired or is invalid")
		}
		return err
	}

	if storedOtp != otp {
		return apperror.NewValidation("Incorrect OTP code")
	}

	s.cache.Delete(ctx, otpKey)

	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.NewNotFound("User not found")
	}

	err = s.userRepository.UpdateUserStatus(ctx, user.ID.String(), UserEntity.StatusActive)
	if err != nil {
		return err
	}

	event := types.Event{
		EventID:       uuid.New().String(),
		EventType:     rabbitmq.OTPVerifiedRK,
		EventVersion:  "1.0.0",
		Timestamp:     time.Now().UTC(),
		Producer:      "go-core-service",
		CorrelationID: uuid.New().String(),
		Payload: map[string]interface{}{
			"email":     user.Email,
			"full_name": user.FullName,
		},
	}

	ctxPub, cancel := context.WithTimeout(ctx, asyncPublishTimeout)
	defer cancel()
	if err := s.publisher.PublishWithContext(ctxPub, event); err != nil {
		log.Printf("[VerifyOTP] failed to publish otp.verified event: %v", err)
	}

	return nil
}

func (s *AuthenService) RefreshToken(ctx context.Context, oldRefreshToken string) (string, string, error) {
	parts := strings.Split(oldRefreshToken, ".")
	if len(parts) != 2 {
		return "", "", apperror.NewUnauthorized("Invalid refresh token format")
	}
	userID := parts[0]
	rawToken := parts[1]

	blacklistKey := fmt.Sprintf("blacklist:refresh:%s", rawToken)
	blacklisted, err := s.cache.Get(ctx, blacklistKey)
	if err == nil && blacklisted == "true" {
		return "", "", apperror.NewUnauthorized("Token has been blacklisted")
	}

	hashedToken := utils.HashToken(rawToken)
	tokenKey := fmt.Sprintf("refresh_token:%s:%s", userID, hashedToken)

	storedUserID, err := s.cache.Get(ctx, tokenKey)
	if err != nil {
		if errors.Is(err, redis.Nil) || err.Error() == "redis: nil" {
			return "", "", apperror.NewUnauthorized("Invalid or expired refresh token")
		}
		return "", "", err
	}

	s.cache.Delete(ctx, tokenKey)

	user, err := s.userRepository.GetUserByID(ctx, storedUserID)
	if err != nil {
		return "", "", err
	}
	if user == nil || user.Status != UserEntity.StatusActive {
		return "", "", apperror.NewUnauthorized("User is inactive or not found")
	}

	newAccessToken, err := utils.GenerateAccessToken(user.ID.String(), user.Email, string(user.Role))
	if err != nil {
		return "", "", err
	}

	newRawToken := uuid.NewString()
	newHashedToken := utils.HashToken(newRawToken)
	newTokenKey := fmt.Sprintf("refresh_token:%s:%s", user.ID.String(), newHashedToken)

	err = s.cache.Set(ctx, newTokenKey, user.ID.String(), 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	newClientRefreshToken := fmt.Sprintf("%s.%s", user.ID.String(), newRawToken)
	return newAccessToken, newClientRefreshToken, nil
}

func (s *AuthenService) Logout(ctx context.Context, refreshToken string) error {
	parts := strings.Split(refreshToken, ".")
	if len(parts) != 2 {
		return apperror.NewValidation("Invalid refresh token format")
	}
	userID := parts[0]
	rawToken := parts[1]

	hashedToken := utils.HashToken(rawToken)
	tokenKey := fmt.Sprintf("refresh_token:%s:%s", userID, hashedToken)

	_ = s.cache.Delete(ctx, tokenKey)

	blacklistKey := fmt.Sprintf("blacklist:refresh:%s", rawToken)
	err := s.cache.Set(ctx, blacklistKey, "true", 7*24*time.Hour)
	if err != nil {
		return err
	}

	return nil
}


func (s *AuthenService) ConsumerOTPEvent(ctx context.Context) error {

	return s.consumer.StartConsumer(&consumer.ConsumerSpec{
		Queue:    "otp.events",
		Prefetch: 10,

		Handler: func(ctx context.Context, msg amqp.Delivery) error {

			var envelope otpEventEnvelope
			if err := json.Unmarshal(msg.Body, &envelope); err != nil {
				return err
			}

			// FIX: use Data.EventType properly
			if envelope.Data.EventType != rabbitmq.UserRegisteredRK {
				return nil
			}

			var payload userRegisteredPayload
			if err := json.Unmarshal(envelope.Data.Payload, &payload); err != nil {
				return err
			}

			if payload.Email == "" {
				return errors.New("missing email")
			}

			otpCode, err := utils.GenerateOTP()
			if err != nil {
				return err
			}

			if err := s.cache.Set(ctx,
				fmt.Sprintf("otp:email:%s", payload.Email),
				otpCode,
				5*time.Minute,
			); err != nil {
				return err
			}

			// OTP → AI EVENT
			event := types.Event{
				EventID:       uuid.New().String(),
				EventType:     rabbitmq.OTPRegisterUserRK,
				EventVersion:  "1.0.0",
				Timestamp:     time.Now().UTC(),
				Producer:      "otp-worker",
				CorrelationID: envelope.Data.CorrelationID,
				Payload: map[string]interface{}{
					"email":     payload.Email,
					"full_name": payload.FullName,
					"phone":     payload.Phone,
					"otp_code":  otpCode,
				},
			}

			ctxPub, cancel := context.WithTimeout(ctx, asyncPublishTimeout)
			defer cancel()

			return s.publisher.PublishWithContext(ctxPub, event)
		},
	})
}
