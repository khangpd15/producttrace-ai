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
// It wraps the event in a NestJS-compatible format (with 'pattern' and 'data' fields).
func (p *Publisher) Publish(event types.Event) error {
	return p.PublishWithContext(context.Background(), event)
}

// PublishWithContext serializes and publishes an event using the provided context.
func (p *Publisher) PublishWithContext(ctx context.Context, event types.Event) error {
	nestMsg := struct {
		Pattern string      `json:"pattern"`
		Data    types.Event `json:"data"`
	}{
		Pattern: event.EventType,
		Data:    event,
	}

	body, err := json.Marshal(nestMsg)
	if err != nil {
		return err
	}

	return p.mgr.Publish(ctx, event.EventType, body)
}
