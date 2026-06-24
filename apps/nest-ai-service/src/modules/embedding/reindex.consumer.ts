import { Controller, Logger } from '@nestjs/common';
import { EventPattern } from '@nestjs/microservices';

import { ReindexService } from './reindex.service';
import { KAFKA } from '../../kafka/kafka.constants';

@Controller()
export class ReindexConsumer {
  private readonly logger =
    new Logger(ReindexConsumer.name);

  constructor(
    private readonly reindexService: ReindexService,
  ) {}

  @EventPattern(
    KAFKA.TOPICS.EMBEDDING_REINDEX_REQUESTED,
  )
  async handleReindex(): Promise<void> {
    this.logger.log(
      '[REINDEX] Request received',
    );

    await this.reindexService.reindexAll();

    this.logger.log(
      '[REINDEX] Completed',
    );
  }
}