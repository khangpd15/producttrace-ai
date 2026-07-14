package adapters

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/rabbitmq"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/types"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/service"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/utils"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/cache"

	"github.com/redis/go-redis/v9"
)

type EventPublisher interface {
	PublishWithContext(ctx context.Context, event types.Event) error
}

type EmailAdapter struct {
	cache     cache.Cache
	publisher EventPublisher
}

func NewEmailAdapter(
	c cache.Cache,
	publisher EventPublisher,
) service.IEmailOTPClient {
	return &EmailAdapter{
		cache:     c,
		publisher: publisher,
	}
}

func (a *EmailAdapter) RequestOTP(
	ctx context.Context,
	email string,
	productIDStr string,
) error {

	otp, err := utils.GenerateOTP()
	if err != nil {
		return err
	}

	key := fmt.Sprintf("otp:ownership:%s", email)

	if err := a.cache.Set(
		ctx,
		key,
		otp,
		5*time.Minute,
	); err != nil {
		return err
	}

	event := types.Event{
		EventID:       uuid.New().String(),
		EventType:     rabbitmq.OTPOwnership,
		EventVersion:  "1.0.0",
		Timestamp:     time.Now().UTC(),
		Producer:      "go-core-service",
		CorrelationID: uuid.New().String(),
		Payload: map[string]interface{}{
			"email":      email,
			"otp_code":   otp,
			"product_id": productIDStr,
		},
	}

	ctxPub, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := a.publisher.PublishWithContext(ctxPub, event); err != nil {
		return err
	}

	return nil
}

func (a *EmailAdapter) ValidateOTP(
	ctx context.Context,
	email string,
	otp string,
) (bool, error) {

	key := fmt.Sprintf("otp:ownership:%s", email)

	storedOTP, err := a.cache.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	if storedOTP != otp {
		return false, nil
	}

	_ = a.cache.Delete(ctx, key)

	return true, nil
}
