import { Injectable, Logger, OnModuleDestroy, OnModuleInit } from '@nestjs/common';
import * as amqp from 'amqplib';
import { RABBITMQ } from './rabbitmq.constants';

/**
 * RabbitMQProducerService — publishes events directly via amqplib.
 *
 * WHY NOT ClientProxy.emit()?
 *   NestJS Transport.RMQ wraps every message in { pattern, data } envelope
 *   and may target the wrong exchange when 'queue' / 'routingKey' options are
 *   not fully configured. Using amqplib.channel.publish() gives us direct,
 *   predictable control over exchange, routing key and message format.
 */
@Injectable()
export class RabbitMQProducerService implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(RabbitMQProducerService.name);

  private connection: amqp.ChannelModel | null = null;
  private channel: amqp.Channel | null = null;
  private isDestroyed = false;

  // ─── Lifecycle ────────────────────────────────────────────────────────────

  async onModuleInit(): Promise<void> {
    await this.connect();
  }

  async onModuleDestroy(): Promise<void> {
    this.isDestroyed = true;
    await this.close();
  }

  // ─── Connect ──────────────────────────────────────────────────────────────

  private async connect(): Promise<void> {
    this.logger.log('[PRODUCER] Connecting to RabbitMQ…');

    this.connection = await amqp.connect(RABBITMQ.URL);
    this.channel = await this.connection.createChannel();

    // Assert the main exchange so we can publish to it immediately.
    await this.channel.assertExchange(
      RABBITMQ.EXCHANGE,
      RABBITMQ.EXCHANGE_TYPE,
      { durable: true },
    );

    this.connection.on('error', (err) => {
      this.logger.error('[PRODUCER] Connection error', err);
    });

    this.connection.on('close', () => {
      if (!this.isDestroyed) {
        this.logger.warn('[PRODUCER] Connection closed — reconnecting in 5s…');
        setTimeout(() => this.connect(), 5_000);
      }
    });

    this.logger.log('[PRODUCER] RabbitMQ producer ready');
  }

  // ─── Publish ──────────────────────────────────────────────────────────────

  /**
   * Publish a message to the main exchange with the given routing key.
   *
   * @param routingKey  e.g. RABBITMQ.ROUTING_KEYS.EMBEDDING_GENERATED
   * @param message     Plain JS object — will be JSON-serialised
   */
  async emit(routingKey: string, message: unknown): Promise<void> {
    if (!this.channel) {
      this.logger.error('[PRODUCER] Channel not ready — cannot publish');
      throw new Error('RabbitMQProducerService: channel not initialised');
    }

    const content = Buffer.from(JSON.stringify(message));

    const published = this.channel.publish(
      RABBITMQ.EXCHANGE,
      routingKey,
      content,
      {
        persistent: true,          // survive broker restart
        contentType: 'application/json',
        timestamp: Math.floor(Date.now() / 1000),
      },
    );

    if (!published) {
      this.logger.warn(
        `[PRODUCER] channel.publish returned false for routingKey=${routingKey} — channel buffer may be full`,
      );
    } else {
      this.logger.log(
        `[PRODUCER] Published → exchange=${RABBITMQ.EXCHANGE} routingKey=${routingKey}`,
      );
    }
  }

  // ─── Shutdown ─────────────────────────────────────────────────────────────

  private async close(): Promise<void> {
    try {
      if (this.channel) {
        await this.channel.close();
        this.channel = null;
      }
      if (this.connection) {
        await this.connection.close();
        this.connection = null;
      }
    } catch (err) {
      this.logger.error('[PRODUCER] Error during close', err);
    }
  }
}