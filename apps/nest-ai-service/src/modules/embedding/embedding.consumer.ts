import { Controller, Logger, OnModuleDestroy, OnModuleInit } from '@nestjs/common';
import * as amqp from 'amqplib';
import { EmbeddingService } from './embedding.service';
import { RABBITMQ } from '../../messaging/rabbitmq/rabbitmq.constants';
import { Event } from '../../messaging/types/event.interface';

@Controller()
export class EmbeddingConsumer implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(EmbeddingConsumer.name);
  private connection: amqp.ChannelModel | null = null;
  private channel: amqp.Channel | null = null;

  constructor(private readonly embeddingService: EmbeddingService) {}

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



    await this.channel.consume(
      RABBITMQ.QUEUES.EMBEDDING,
      async (message) => {
        if (!message) {
          return;
        }

        try {
          const event = this.normalizeEvent(message.content);
          if (!event) {
            this.logger.error('Invalid event payload received');
            this.channel?.nack(message, false, false);
            return;
          }

          this.logger.log(
            `Received event id=${event.event_id ?? 'unknown'} type=${event.event_type ?? 'unknown'}`,
          );

          await this.embeddingService.processEvent(event);
          this.channel?.ack(message);
        } catch (error) {
          this.logger.error(
            `Embedding consumer failed: ${error instanceof Error ? error.message : JSON.stringify(error)}`,
          );
          this.channel?.nack(message, false, false);
        }
      },
      { noAck: false },
    );

    this.logger.log(`Embedding consumer listening on ${RABBITMQ.QUEUES.EMBEDDING}`);
  }

  private normalizeEvent(payload: unknown): Event | null {
    let parsed: unknown = payload;

    if (Buffer.isBuffer(parsed)) {
      parsed = parsed.toString('utf8');
    }

    if (typeof parsed === 'string') {
      const stripped = parsed.trim();
      if (!stripped) {
        this.logger.error('Empty payload received');
        return null;
      }

      try {
        parsed = JSON.parse(stripped);
      } catch (error) {
        this.logger.warn(`Treating payload as plain text: ${stripped}`);
        try {
          parsed = this.parseLooseObject(stripped);
        } catch (looseError) {
          this.logger.error(`Invalid JSON payload: ${stripped}`);
          return null;
        }
      }
    }

    if (
      parsed &&
      typeof parsed === 'object' &&
      'pattern' in parsed &&
      'data' in parsed &&
      typeof (parsed as { data?: unknown }).data === 'object'
    ) {
      parsed = (parsed as { data: unknown }).data;
    }

    if (
      parsed &&
      typeof parsed === 'object' &&
      'event_id' in parsed &&
      'event_type' in parsed
    ) {
      return parsed as Event;
    }

    this.logger.warn(`Unsupported event payload shape: ${JSON.stringify(parsed)}`);
    return null;
  }

  private parseLooseObject(input: string): Record<string, unknown> {
    const normalized = input
      .replace(/([{,]\s*)([A-Za-z0-9_.$-]+)(\s*:)/g, '$1"$2"$3')
      .replace(/'([^']*)'/g, '"$1"');

    return JSON.parse(normalized);
  }
}