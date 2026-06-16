package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"


func SetupTopology(ch *amqp.Channel) error {
    // 1. Tạo exchange chính
    if err := ch.ExchangeDeclare(
        EventExchange,
        "topic",
        true,
        false,
        false,
        false,
        nil,
    ); err != nil {
        return err
    }

    // 2. Tạo DLX
    if err := ch.ExchangeDeclare(
        DLXExchange,
        "direct",
        true,
        false,
        false,
        false,
        nil,
    ); err != nil {
        return err
    }

    // 3. Tạo queue
    args := amqp.Table{
        "x-dead-letter-exchange":    DLXExchange,
        "x-dead-letter-routing-key": AIDLQRoutingKey,
    }

    if _, err := ch.QueueDeclare(
        AIQueue,
        true,
        false,
        false,
        false,
        args,
    ); err != nil {
        return err
    }

    // 4. Tạo DLQ
    if _, err := ch.QueueDeclare(
        AIDLQ,
        true,
        false,
        false,
        false,
        nil,
    ); err != nil {
        return err
    }

    // 5. Bind queue chính
    if err := ch.QueueBind(
        AIQueue,
        ProductCreatedRK,
        EventExchange,
        false,
        nil,
    ); err != nil {
        return err
    }

    // 6. Bind DLQ
    if err := ch.QueueBind(
        AIDLQ,
        AIDLQRoutingKey,
        DLXExchange,
        false,
        nil,
    ); err != nil {
        return err
    }

    return nil
}