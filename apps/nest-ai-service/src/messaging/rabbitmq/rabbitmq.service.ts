import {
  Injectable,
  OnModuleInit,
  OnModuleDestroy,
  Logger,
} from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import * as amqp from 'amqplib';
import * as fs from 'fs';

import { RABBITMQ } from './rabbitmq.constants';
import { NotificationConsumer } from '../consumers/notification.consumer';

@Injectable()
export class RabbitMQService implements OnModuleInit, OnModuleDestroy {

  private readonly logger = new Logger(RabbitMQService.name);

  private connection: amqp.ChannelModel | null = null;
  private channel: amqp.Channel | null = null;

  private isConnecting = false;
  private isDestroyed = false;

  constructor(
    private readonly configService: ConfigService,
    private readonly notificationConsumer: NotificationConsumer,
  ) { }

  // ─── Lifecycle ────────────────────────────────────────────────────────────

  async onModuleInit() {
    await this.connect();
  }

  async onModuleDestroy() {
    this.isDestroyed = true;
    await this.close();
  }

  // ─── Connection ───────────────────────────────────────────────────────────

  private async connect() {
    if (this.isConnecting || this.connection || this.isDestroyed) return;

    this.isConnecting = true;

    const isDocker = fs.existsSync('/.dockerenv');
    const envUrl = this.configService.get<string>('RABBITMQ_URL');
    const defaultUrl = isDocker
      ? 'amqp://admin:admin123@rabbitmq:5672/%2F'
      : 'amqp://admin:admin123@localhost:5672/%2F';
    const url = !isDocker && envUrl?.includes('rabbitmq')
      ? defaultUrl
      : envUrl || RABBITMQ.URL || defaultUrl;

    this.logger.log(`Connecting to RabbitMQ ${url.replace(/:([^:@]+)@/, ':****@')}`);

    try {
      this.connection = await amqp.connect(url);
      this.logger.log('RabbitMQ connected');

      this.connection.on('error', err =>
        this.logger.error('RabbitMQ connection error', err),
      );

      this.connection.on('close', () => {
        this.logger.warn('RabbitMQ connection closed');
        this.handleDisconnect();
      });

      await this.createChannel();
      this.isConnecting = false;

    } catch (error) {
      this.logger.error('RabbitMQ connection failed', error);
      await this.close();
      this.isConnecting = false;
      this.handleDisconnect();
    }
  }

  // ─── Channel ──────────────────────────────────────────────────────────────

  private async createChannel() {
    if (!this.connection) return;

    try {
      this.channel = await this.connection.createChannel();

      this.channel.on('error', err =>
        this.logger.error('RabbitMQ channel error', err),
      );

      this.channel.on('close', () =>
        this.logger.warn('RabbitMQ channel closed'),
      );

      // Set prefetch: process one message at a time per consumer
      await this.channel.prefetch(10);

      await this.setupTopology();

    } catch (error) {
      this.logger.error('Channel creation failed', error);
      throw error;
    }
  }

  // ─── Topology ─────────────────────────────────────────────────────────────

  /**
   * Asserts the exchange/queue topology and starts the single NotificationConsumer.
   *
   * NOTE: Go Core Service owns and already declares these resources.
   *       NestJS re-asserts them (idempotent in RabbitMQ) to ensure the channel
   *       fails fast if topology drifts, then adds its routing-key bindings.
   */
  private async setupTopology() {
    if (!this.channel) return;

    const ch = this.channel;
    const queueName = RABBITMQ.QUEUES.NOTIFICATION;     // "ai.events"
    const failedQueue = `${queueName}.failed`;             // "ai.events.failed"

    // 1. Assert main exchange (idempotent — must match Go's declaration)
    await ch.assertExchange(RABBITMQ.EXCHANGE, RABBITMQ.EXCHANGE_TYPE, { durable: true });
    this.logger.log(`Exchange asserted: ${RABBITMQ.EXCHANGE}`);

    // 2. Assert DLX (idempotent)
    await ch.assertExchange(RABBITMQ.DLX_EXCHANGE, 'topic', { durable: true });
    this.logger.log(`DLX exchange asserted: ${RABBITMQ.DLX_EXCHANGE}`);

    // 3. Assert DLQ (idempotent)
    await ch.assertQueue(failedQueue, { durable: true });
    await ch.bindQueue(failedQueue, RABBITMQ.DLX_EXCHANGE, failedQueue);
    this.logger.log(`DLQ asserted and bound: ${failedQueue}`);

    // 4. Assert main notification queue (parameters must match Go's declaration)
    await ch.assertQueue(queueName, {
      durable: true,
      arguments: {
        'x-dead-letter-exchange': RABBITMQ.DLX_EXCHANGE,
        'x-dead-letter-routing-key': failedQueue,
      },
    });
    this.logger.log(`Queue asserted: ${queueName}`);

    // 5. Bind all routing keys → ai.events (idempotent; Go may already bind some)
    const routingKeys = [
      RABBITMQ.ROUTING_KEYS.USER_REGISTERED, // "otp.registered"
      RABBITMQ.ROUTING_KEYS.PASSWORD_RESET,  // "otp.forgot"
      RABBITMQ.ROUTING_KEYS.USER_VERIFIED,   // "otp.verified"
      RABBITMQ.ROUTING_KEYS.PRODUCT_CREATED, // "product.created"
      RABBITMQ.ROUTING_KEYS.NOTIFICATION_SENT, // "notification.sent"
      RABBITMQ.ROUTING_KEYS.OWNERSHIP_OTP,
      RABBITMQ.ROUTING_KEYS.OWNERSHIP_TRANSFERRED,
    ];

    for (const rk of routingKeys) {
      await ch.bindQueue(queueName, RABBITMQ.EXCHANGE, rk);
      this.logger.log(`Bound: ${queueName} ← [${RABBITMQ.EXCHANGE}] ${rk}`);
    }

    // 6. Start the single unified consumer
    await ch.consume(
      queueName,
      async (msg) => {
        if (!msg) return;
        try {
          await this.notificationConsumer.handleMessage(msg, ch);
        } catch (error) {
          this.logger.error(`Unhandled consumer error on ${queueName}`, error);
        }
      },
      { noAck: false },
    );

    this.logger.log(`NotificationConsumer started on queue: ${queueName}`);
  }

  // ─── Reconnect ────────────────────────────────────────────────────────────

  private handleDisconnect() {
    if (this.isDestroyed) return;

    this.connection = null;
    this.channel = null;

    setTimeout(() => this.connect(), 5000);
  }

  // ─── Shutdown ─────────────────────────────────────────────────────────────

  private async close() {
    try {
      if (this.channel) {
        await this.channel.close();
        this.channel = null;
      }
      if (this.connection) {
        await this.connection.close();
        this.connection = null;
      }
    } catch (error) {
      this.logger.error('RabbitMQ close error', error);
    }
  }
}