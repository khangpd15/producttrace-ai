package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// SetupTopology declares the necessary topic exchanges, queues, and bindings.
// It ensures that queues and bindings exist before any messages are published.
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

	// Declare the main queue for NestJS notification consumer
	q, err := ch.QueueDeclare(
		"ai.events", // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    DLXExchange,
			"x-dead-letter-routing-key": "ai.events.failed",
		},
	)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}
	routingNestKeys := []string{
		OTPRegisterUserRK,
		OTPForgotRK,
		OTPVerifiedRK,
		ProductCreatedRK,
		TraceExportedRK,
	}

	for _, rk := range routingNestKeys {
		err = ch.QueueBind(
			q.Name,
			rk,
			DefaultExchange,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("bind queue %s to exchange %s with rk %s: %w", q.Name, DefaultExchange, rk, err)
		}
	}

	// Declare Embedding queue and bind it to the product-created and trace events
	embeddingQueueArgs := amqp.Table{
		"x-dead-letter-exchange":    EmbeddingDLXExchange,
		"x-dead-letter-routing-key": EmbeddingDLQName,
	}
	embeddingQueue, err := ch.QueueDeclare(
		EmbeddingQueueName,
		true,
		false,
		false,
		false,
		embeddingQueueArgs,
	)
	if err != nil {
		return fmt.Errorf("declare embedding queue: %w", err)
	}

	for _, rk := range []string{ProductCreatedRK, TraceExportedRK} {
		err = ch.QueueBind(
			embeddingQueue.Name,
			rk,
			DefaultExchange,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("bind embedding queue %s to exchange %s with rk %s: %w", embeddingQueue.Name, DefaultExchange, rk, err)
		}
	}

	// Declare Embedding DLX exchange and DLQ queue
	err = ch.ExchangeDeclare(
		EmbeddingDLXExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare embedding dlx exchange %s: %w", EmbeddingDLXExchange, err)
	}

	embeddingDLQ, err := ch.QueueDeclare(
		EmbeddingDLQName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare embedding dlq: %w", err)
	}

	err = ch.QueueBind(
		embeddingDLQ.Name,
		EmbeddingDLQName,
		EmbeddingDLXExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("bind embedding dlq: %w", err)
	}

	// Declare DLQ queue and bind it
	dlq, err := ch.QueueDeclare(
		"ai.events.failed", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare dlq: %w", err)
	}

	err = ch.QueueBind(
		dlq.Name,
		"ai.events.failed",
		DLXExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("bind dlq: %w", err)
	}

	// Declare notification.password.reset queue for NestJS PasswordResetConsumer

	qOtp, err := ch.QueueDeclare(
		"otp.events", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    DLXExchange,
			"x-dead-letter-routing-key": "otp.events.failed",
		},
	)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	routingOTPKeys := []string{
		UserRegisteredRK,
		UserPasswordForgotRK,
	}
	for _, rk := range routingOTPKeys {
		err = ch.QueueBind(
			qOtp.Name,
			rk,
			DefaultExchange,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("bind queue %s to exchange %s with rk %s: %w", qOtp.Name, DefaultExchange, rk, err)
		}
	}
	dlqOTP, err := ch.QueueDeclare(
		"otp.events.failed", // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare dlq: %w", err)
	}
	err = ch.QueueBind(
		dlqOTP.Name,
		"otp.events.failed",
		DLXExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("bind dlq: %w", err)
	}

	qBatch, err := ch.QueueDeclare(
		"batch.events", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    DLXExchange,
			"x-dead-letter-routing-key": "batch.events.failed",
		},
	)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	routingGoKeys := []string{
		BatchCreatedRK,
		BatchDeletedRK,
		BatchUpdatedRK,
	}
	for _, rk := range routingGoKeys {
		err = ch.QueueBind(
			qBatch.Name,
			rk,
			DefaultExchange,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("bind queue %s to exchange %s with rk %s: %w", qBatch.Name, DefaultExchange, rk, err)
		}
	}
	dlqBatch, err := ch.QueueDeclare(
		"batch.events.failed", // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare dlq: %w", err)
	}
	err = ch.QueueBind(
		dlqBatch.Name,
		"batch.events.failed",
		DLXExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("bind dlq: %w", err)
	}

	// Bind routing keys

	return nil
}
