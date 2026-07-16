import { Controller, Post, HttpCode, HttpStatus, Logger } from '@nestjs/common';
import { ReindexService } from './reindex.service';

/**
 * ReindexController — exposes an HTTP endpoint to trigger a full reindex
 * without needing a RabbitMQ event. Useful for manual testing and admin ops.
 *
 * POST /admin/reindex
 *   Returns 202 Accepted immediately; reindex runs asynchronously in background.
 */
@Controller('admin')
export class ReindexController {
  private readonly logger = new Logger(ReindexController.name);

  constructor(private readonly reindexService: ReindexService) {}

  @Post('reindex')
  @HttpCode(HttpStatus.ACCEPTED)
  async triggerReindex() {
    this.logger.log('[REINDEX] HTTP trigger received — starting reindex in background…');

    // Fire-and-forget: do NOT await, so the HTTP response returns immediately
    // while reindex runs asynchronously.
    this.reindexService.reindexAll().catch((err: unknown) => {
      this.logger.error(
        `[REINDEX] Background reindex failed: ${err instanceof Error ? err.message : JSON.stringify(err)}`,
      );
    });

    return {
      message: 'Reindex started',
      timestamp: new Date().toISOString(),
    };
  }
}
