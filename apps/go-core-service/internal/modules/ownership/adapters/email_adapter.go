package adapters

import (
	"context"
	"log"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/service"
)

type DummyEmailAdapter struct{}

func NewDummyEmailAdapter() service.IEmailOTPClient {
	return &DummyEmailAdapter{}
}

func (a *DummyEmailAdapter) RequestOTP(ctx context.Context, email string, productIDStr string) error {
	log.Printf("[EmailAdapter] Faking RequestOTP for Email: %s, Product: %s\n", email, productIDStr)
	return nil
}

func (a *DummyEmailAdapter) ValidateOTP(ctx context.Context, email string, otp string) (bool, error) {
	log.Printf("[EmailAdapter] Faking ValidateOTP for Email: %s, OTP: %s\n", email, otp)
	// Hardcoded dummy valid OTP check
	return otp == "123456", nil
}
