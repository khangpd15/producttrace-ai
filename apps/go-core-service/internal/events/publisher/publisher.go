package publisher

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/types"
)

type Publisher struct {
	channel  *amqp.Channel
	exchange string
}

func New(ch *amqp.Channel, exchange string) *Publisher {
	return &Publisher{
		channel:  ch,
		exchange: exchange,
	}
}

func (p *Publisher) Publish(event types.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(
		context.Background(),
		p.exchange,
		event.EventType,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
