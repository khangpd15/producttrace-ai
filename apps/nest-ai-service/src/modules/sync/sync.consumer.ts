import { Controller, Logger, OnModuleDestroy, OnModuleInit } from "@nestjs/common";
import * as amqp from "amqplib";

import { SyncService } from "./sync.service";
import { RABBITMQ } from "../../messaging/rabbitmq/rabbitmq.constants";

@Controller()
export class SyncConsumer implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(SyncConsumer.name);
  private connection: amqp.ChannelModel | null = null;
  private channel: amqp.Channel | null = null;

  constructor(private readonly syncService: SyncService) {}

  async onModuleInit(): Promise<void> {
    await this.connectAndConsume();
  }

  async onModuleDestroy(): Promise<void> {
    await this.channel?.close();
    await this.connection?.close();
  }

  private async connectAndConsume(): Promise<void> {
    this.connection = await amqp.connect(RABBITMQ.URL);
    this.channel = await this.connection.createChannel();

    await this.channel.assertExchange(RABBITMQ.EXCHANGE, RABBITMQ.EXCHANGE_TYPE, {
      durable: true,
    });

    await this.channel.assertQueue(RABBITMQ.QUEUES.EMBEDDING_SYNC, {
      durable: true,
      arguments: {
        'x-dead-letter-exchange': RABBITMQ.DLX.EMBEDDING,
        'x-dead-letter-routing-key': RABBITMQ.DLQ_ROUTING_KEYS.EMBEDDING,
      },
    });

    await this.channel.bindQueue(
      RABBITMQ.QUEUES.EMBEDDING_SYNC,
      RABBITMQ.EXCHANGE,
      RABBITMQ.ROUTING_KEYS.EMBEDDING_GENERATED,
    );

    await this.channel.consume(
      RABBITMQ.QUEUES.EMBEDDING_SYNC,
      async (message) => {
        if (!message) {
          return;
        }

        try {
          const event = this.normalizeEvent(message.content);
          if (!event) {
            this.logger.error('Empty event received');
            this.channel?.nack(message, false, false);
            return;
          }

          this.logger.log(
            `Received rabbitmq event id=${event?.event_id ?? 'unknown'} type=${event?.event_type ?? 'unknown'}`,
          );

          await this.syncService.process(event);
          this.logger.log(`[SYNC] SUCCESS point=${event?.pointId ?? 'unknown'}`);
          this.channel?.ack(message);
        } catch (error) {
          const errMsg = error instanceof Error ? error.message : JSON.stringify(error);
          this.logger.error(`[SYNC] FAILED id=${(event as any)?.event_id ?? 'unknown'} error=${errMsg}`);
          this.channel?.nack(message, false, true);
        }
      },
      { noAck: false },
    );

    this.logger.log(`Sync consumer listening on ${RABBITMQ.QUEUES.EMBEDDING_SYNC}`);
  }

  private normalizeEvent(payload: unknown): any {
    let data = payload;

    if (Buffer.isBuffer(data)) {
      data = data.toString('utf8');
    }

    if (typeof data === 'string') {
      try {
        data = JSON.parse(data);
      } catch (error) {
        this.logger.error('Invalid JSON payload');
        return null;
      }
    }

    if (
      data &&
      typeof data === 'object' &&
      'pattern' in data &&
      'data' in data &&
      typeof (data as { data?: unknown }).data === 'object'
    ) {
      return (data as { data: unknown }).data;
    }

    return data;
  }
}