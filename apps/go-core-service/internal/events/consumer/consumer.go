package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type HandlerFunc func(ctx context.Context, msg amqp.Delivery) error

type ConsumerSpec struct {
	Queue    string
	Prefetch int
	Handler  HandlerFunc
}

type Consumer struct {
	manager *rabbitmq.Manager
}

func NewConsumer(manager *rabbitmq.Manager) *Consumer {
	return &Consumer{manager: manager}
}

func (c *Consumer) StartConsumer(spec *ConsumerSpec) error {
	conn := c.manager.ChannelConnection()
	if conn == nil {
		return fmt.Errorf("rabbitmq not connected")
	}

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open consumer channel: %w", err)
	}

	prefetch := spec.Prefetch
	if prefetch <= 0 {
		prefetch = 10
	}

	if err := ch.Qos(prefetch, 0, false); err != nil {
		_ = ch.Close()
		return err
	}

	msgs, err := ch.Consume(
		spec.Queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		return err
	}

	// SAFE DISPATCH (NO GOROUTINE EXPLOSION)
	go c.dispatch(spec, msgs, ch)

	return nil
}

func (c *Consumer) dispatch(spec *ConsumerSpec, msgs <-chan amqp.Delivery, ch *amqp.Channel) {
	for msg := range msgs {

		ctx, cancel := context.WithTimeout(c.manager.Context(), 30*time.Second)

		err := spec.Handler(ctx, msg)

		if err != nil {
			_ = msg.Nack(false, true)
			cancel()
			continue
		}

		_ = msg.Ack(false)
		cancel()
	}

	_ = ch.Close()
}
