import { Injectable, Logger } from '@nestjs/common';
import { EmbeddingService } from '../embedding/embedding.service';
import { ProductClientService } from '../../integrations/go-core/product-client.service';

@Injectable()
export class ReindexService {
  private readonly logger = new Logger(ReindexService.name);

  constructor(
    private readonly embeddingService: EmbeddingService,
    private readonly productClient: ProductClientService,
  ) { }

  async reindexAll(): Promise<void> {
    this.logger.log('[REINDEX][START]');

    let page = 1;
    const limit = 100;

    while (true) {
      const result =
        await this.productClient.getProducts(page, limit);

      this.logger.log(
        `[REINDEX][PRODUCT_COUNT] page=${page} products=${result.data.length}`,
      );

      if (result.data.length === 0) {
        break;
      }

      for (const product of result.data) {
        // Use a timestamp suffix so that each reindex run generates a fresh
        // event_id — preventing EmbeddingService.processedEvents from
        // skipping duplicates on the second (and subsequent) reindex runs.
        const runId = Math.floor(Date.now() / 1000); // second-precision is enough

        await this.embeddingService.processEvent({
          event_id: `reindex-${product.id}-${runId}`,
          event_type: 'product.reindexed',
          event_version: '1.0',
          producer: 'nest-ai-service',
          correlation_id: `reindex-${product.id}-${runId}`,
          timestamp: new Date().toISOString(),
          payload: {
            productId: product.id,
            category: product.category_id,
            name: product.name,
            description: product.description,
            status: product.status,
          },
        });
      }

      this.logger.log(
        `Reindexed page=${page}, products=${result.data.length}`,
      );

      if (page >= result.total_pages) {
        break;
      }

      page++;
    }

    this.logger.log('[REINDEX][COMPLETED]');
  }
}
