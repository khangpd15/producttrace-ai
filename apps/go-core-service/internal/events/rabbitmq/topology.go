package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// SetupTopology declares exchanges, queues, and bindings
// Publisher chỉ tạo exchange, nhưng ở đây bạn đang setup full topology (dev-friendly)
func SetupTopology(ch *amqp.Channel) error {

	// 1. Declare exchanges
	exchanges := []struct {
		name string
		kind string
	}{
		{EventExchange, "topic"},
		{DLXExchange, "direct"},
	}

	for _, ex := range exchanges {
		if err := ch.ExchangeDeclare(
			ex.name,
			ex.kind,
			true,  // durable
			false, // auto-deleted
			false, // internal
			false, // no-wait
			nil,
		); err != nil {
			return fmt.Errorf("declare exchange %s: %w", ex.name, err)
		}
	}

	// 2. Queue arguments (DLX setup)
	args := amqp.Table{
		"x-dead-letter-exchange":    DLXExchange,
		"x-dead-letter-routing-key": AIDLQRoutingKey,
	}

	// 3. Declare main queue
	if _, err := ch.QueueDeclare(
		AIQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,
	); err != nil {
		return fmt.Errorf("declare AIQueue: %w", err)
	}

	// 4. Declare DLQ
	if _, err := ch.QueueDeclare(
		AIDLQ,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare AIDLQ: %w", err)
	}

	// 5. Bind main queue to event exchange
	if err := ch.QueueBind(
		AIQueue,
		ProductCreatedRK,
		EventExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind AIQueue: %w", err)
	}

	// 6. Bind DLQ to DLX
	if err := ch.QueueBind(
		AIDLQ,
		AIDLQRoutingKey,
		DLXExchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind AIDLQ: %w", err)
	}

	return nil
}
