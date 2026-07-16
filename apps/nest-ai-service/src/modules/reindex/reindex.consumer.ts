import { Controller, Logger, OnModuleDestroy, OnModuleInit } from '@nestjs/common';
import * as amqp from 'amqplib';

import { ReindexService } from './reindex.service';
import { RABBITMQ } from '../../messaging/rabbitmq/rabbitmq.constants';

/**
 * ReindexConsumer — listens on 'embedding_reindex_queue' for
 * 'embedding.reindex.requested' events and triggers a full reindex.
 *
 * WHY NOT @EventPattern?
 *   @EventPattern / @MessagePattern only work when the app is bootstrapped
 *   with NestFactory.createMicroservice(). This app uses NestFactory.create()
 *   (HTTP server), so those decorators are silently ignored. Using amqplib
 *   directly (same pattern as EmbeddingConsumer and SyncConsumer) ensures the
 *   consumer actually starts.
 */
@Controller()
export class ReindexConsumer implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(ReindexConsumer.name);

  private connection: amqp.ChannelModel | null = null;
  private channel: amqp.Channel | null = null;

  constructor(private readonly reindexService: ReindexService) {}

  // ─── Lifecycle ────────────────────────────────────────────────────────────

  async onModuleInit(): Promise<void> {
    await this.connectAndConsume();
  }

  async onModuleDestroy(): Promise<void> {
    await this.channel?.close();
    await this.connection?.close();
  }

  // ─── Consumer setup ───────────────────────────────────────────────────────

  private async connectAndConsume(): Promise<void> {
    this.connection = await amqp.connect(RABBITMQ.URL);
    this.channel = await this.connection.createChannel();


    // Only one reindex job at a time
    await this.channel.prefetch(1);

    await this.channel.consume(
      RABBITMQ.QUEUES.EMBEDDING_REINDEX,
      async (message) => {
        if (!message) return;

        this.logger.log('[REINDEX] RabbitMQ trigger received — starting reindex…');

        try {
          await this.reindexService.reindexAll();

          this.channel?.ack(message);
          this.logger.log('[REINDEX] Completed successfully');
        } catch (error) {
          this.logger.error(
            `[REINDEX] Failed: ${error instanceof Error ? error.message : JSON.stringify(error)}`,
          );

          // nack without requeue — avoid infinite reindex loop on persistent errors
          this.channel?.nack(message, false, false);
        }
      },
      { noAck: false },
    );

    this.logger.log(
      `[REINDEX] Consumer listening on ${RABBITMQ.QUEUES.EMBEDDING_REINDEX} ← ${RABBITMQ.ROUTING_KEYS.EMBEDDING_REINDEX_REQUESTED}`,
    );
  }
}
