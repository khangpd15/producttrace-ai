package rabbitmq

import (
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Connect establishes a connection to RabbitMQ with a default heartbeat configuration and timeout.
func Connect(url string) (*amqp.Connection, error) {
	cfg := amqp.Config{
		Heartbeat: 10 * time.Second,
		Dial:      amqp.DefaultDial(5 * time.Second),
	}

	conn, err := amqp.DialConfig(url, cfg)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	return conn, nil
}