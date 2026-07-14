import { Controller, Logger } from '@nestjs/common';
import { EventPattern } from '@nestjs/microservices';

import { ReindexService } from './reindex.service';
import { RABBITMQ } from '../../messaging/rabbitmq/rabbitmq.constants';

@Controller()
export class ReindexConsumer {
  private readonly logger =
    new Logger(ReindexConsumer.name);

  constructor(
    private readonly reindexService: ReindexService,
  ) {}

  @EventPattern(
    RABBITMQ.ROUTING_KEYS.EMBEDDING_REINDEX_REQUESTED,
  )
  async handleReindex(): Promise<void> {
    this.logger.log(
      '[REINDEX] RabbitMQ request received',
    );

    try {
      await this.reindexService.reindexAll();

      this.logger.log(
        '[REINDEX] Completed',
      );
    } catch (error) {
      this.logger.error(
        `[REINDEX] Failed: ${
          error instanceof Error
            ? error.message
            : JSON.stringify(error)
        }`,
      );

      throw error;
    }
  }
}