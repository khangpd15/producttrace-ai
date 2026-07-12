package service

import (
	"context"
)

// IEmailOTPClient acts as an anti-corruption layer interface for calling the Mail (NestJS) Service
type IEmailOTPClient interface {
	// RequestOTP calls the external service to generate and send an OTP to the email.
	RequestOTP(ctx context.Context, email string, productIDStr string) error

	// ValidateOTP calls the external service to check if the given OTP is valid.
	ValidateOTP(ctx context.Context, email string, otp string) (bool, error)
}
