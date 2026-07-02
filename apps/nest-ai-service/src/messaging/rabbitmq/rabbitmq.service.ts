import { Injectable, OnModuleInit, OnModuleDestroy, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import * as amqp from 'amqplib';
import { RABBITMQ } from './rabbitmq.constants';
import { UserRegisteredConsumer } from '../consumers/user-registered.consumer';
import { UserVerifiedConsumer } from '../consumers/user-verified.consumer';
import { PasswordResetConsumer } from '../consumers/password-reset.consumer';
import { ProductCreatedConsumer } from '../consumers/product-created.consumer';

@Injectable()
export class RabbitMQService implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(RabbitMQService.name);
  private connection: amqp.ChannelModel | null = null;
  private channel: amqp.Channel | null = null;
  private isConnecting = false;
  private isDestroyed = false;

  constructor(
    private readonly configService: ConfigService,
    private readonly userRegisteredConsumer: UserRegisteredConsumer,
    private readonly userVerifiedConsumer: UserVerifiedConsumer,
    private readonly passwordResetConsumer: PasswordResetConsumer,
    private readonly productCreatedConsumer: ProductCreatedConsumer,
  ) {}

  /**
   * Automatically connect to RabbitMQ when the module initializes.
   */
  async onModuleInit() {
    await this.connect();
  }

  /**
   * Gracefully close connections when the module is destroyed.
   */
  async onModuleDestroy() {
    this.isDestroyed = true;
    await this.close();
  }

  /**
   * Establishes a connection to RabbitMQ with error handling and auto-reconnect.
   */
  private async connect() {
    if (this.isConnecting || this.connection) return;
    this.isConnecting = true;

    const url = this.configService.get<string>('RABBITMQ_URL') || RABBITMQ.URL;
    // Hide password for safety when logging the URL
    const maskedUrl = url.replace(/:([^:@]+)@/, ':****@');
    this.logger.log(`Connecting to RabbitMQ at ${maskedUrl}...`);

    try {
      this.connection = await amqp.connect(url);
      this.logger.log('Successfully connected to RabbitMQ.');

      this.connection.on('error', (err: any) => {
        this.logger.error('RabbitMQ connection error event triggered', err);
      });

      this.connection.on('close', (err: any) => {
        this.logger.warn('RabbitMQ connection closed. Initiating reconnection...', err);
        this.handleDisconnect();
      });

      await this.createChannel();
      this.isConnecting = false;
    } catch (error) {
      this.logger.error('Failed to establish RabbitMQ connection', error);
      this.isConnecting = false;
      this.handleDisconnect();
    }
  }

  /**
   * Creates a communication channel on top of the established connection.
   */
  private async createChannel() {
    if (!this.connection) return;

    try {
      this.channel = await this.connection.createChannel();
      this.logger.log('RabbitMQ channel created successfully.');

      this.channel.on('error', (err: any) => {
        this.logger.error('RabbitMQ channel error event triggered', err);
      });

      this.channel.on('close', () => {
        this.logger.warn('RabbitMQ channel closed.');
      });

      // Declare exchange, queues, and bindings, then start consuming
      await this.setupTopology();
    } catch (error) {
      this.logger.error('Failed to create RabbitMQ channel', error);
      throw error;
    }
  }

  /**
   * Automatically declares exchange, queues, bindings, and attaches consumers.
   */
  private async setupTopology() {
    if (!this.channel) return;

    const exchange = RABBITMQ.EXCHANGE;
    const exchangeType = RABBITMQ.EXCHANGE_TYPE;

    this.logger.log(`Asserting exchange: ${exchange} (type: ${exchangeType})`);
    await this.channel.assertExchange(exchange, exchangeType, { durable: true });

    // Define topology configuration mapping queues, routing keys, and consumers
    const topology = [
      {
        name: RABBITMQ.QUEUES.USER_REGISTERED,
        routingKey: RABBITMQ.ROUTING_KEYS.USER_REGISTERED,
        consumer: this.userRegisteredConsumer,
      },
      {
        name: RABBITMQ.QUEUES.USER_VERIFIED,
        routingKey: RABBITMQ.ROUTING_KEYS.USER_VERIFIED,
        consumer: this.userVerifiedConsumer,
      },
      {
        name: RABBITMQ.QUEUES.PASSWORD_RESET,
        routingKey: RABBITMQ.ROUTING_KEYS.PASSWORD_RESET,
        consumer: this.passwordResetConsumer,
      },
      {
        name: RABBITMQ.QUEUES.PRODUCT_CREATED,
        routingKey: RABBITMQ.ROUTING_KEYS.PRODUCT_CREATED,
        consumer: this.productCreatedConsumer,
      },
    ];

    for (const item of topology) {
      this.logger.log(`Asserting queue: ${item.name}`);
      await this.channel.assertQueue(item.name, { durable: true });

      this.logger.log(`Binding queue ${item.name} to exchange ${exchange} with routing key ${item.routingKey}`);
      await this.channel.bindQueue(item.name, exchange, item.routingKey);

      this.logger.log(`Registering consumer for queue: ${item.name}`);
      await this.channel.consume(
        item.name,
        async (msg) => {
          if (!msg) {
            this.logger.warn(`Received null message on queue: ${item.name}`);
            return;
          }
          try {
            await item.consumer.handleMessage(msg, this.channel!);
          } catch (consumeError) {
            this.logger.error(`Error dispatched to consumer for queue ${item.name}`, consumeError);
          }
        },
        { noAck: false } // Force manual acknowledgment (ACK)
      );
    }
  }

  /**
   * Resets connections and schedules a reconnection attempt with a 5-second delay.
   */
  private handleDisconnect() {
    if (this.isDestroyed) return;

    this.connection = null;
    this.channel = null;

    this.logger.log('Scheduling RabbitMQ reconnection in 5 seconds...');
    setTimeout(() => {
      this.connect();
    }, 5000);
  }

  /**
   * Closes the active channel and connection safely.
   */
  private async close() {
    try {
      if (this.channel) {
        await this.channel.close();
      }
      if (this.connection) {
        await this.connection.close();
      }
    } catch (error) {
      this.logger.error('Error occurred while closing RabbitMQ connection/channel', error);
    }
  }
}
