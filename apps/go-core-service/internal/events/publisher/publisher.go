package publisher

import (
	"context"
	"encoding/json"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/types"
)

// RabbitMQPublisher defines the interface required to publish messages.
type RabbitMQPublisher interface {
	Publish(ctx context.Context, routingKey string, body []byte) error
}

type Publisher struct {
	mgr RabbitMQPublisher
}

// New creates a new Publisher using a RabbitMQ manager.
func New(mgr RabbitMQPublisher) *Publisher {
	return &Publisher{
		mgr: mgr,
	}
}

// Publish serializes and publishes an event to RabbitMQ via the manager.
func (p *Publisher) Publish(event types.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Publish via manager which handles retries, thread safety and confirmations.
	return p.mgr.Publish(context.Background(), event.EventType, body)
}