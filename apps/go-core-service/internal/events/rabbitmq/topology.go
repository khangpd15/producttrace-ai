package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// SetupTopology declares the necessary topic exchanges for the publisher.
// In accordance with event-driven best practices, the publisher is NOT responsible
// for declaring queues or DLQs (which belong to the consumer).
func SetupTopology(ch *amqp.Channel) error {
	exchanges := []string{
		DefaultExchange,
		DLXExchange,
	}

	for _, exchange := range exchanges {
		err := ch.ExchangeDeclare(
			exchange,
			"topic", // type
			true,    // durable
			false,   // auto-deleted
			false,   // internal
			false,   // no-wait
			nil,     // arguments
		)
		if err != nil {
			return fmt.Errorf("declare exchange %s: %w", exchange, err)
		}
	}

	return nil
}
