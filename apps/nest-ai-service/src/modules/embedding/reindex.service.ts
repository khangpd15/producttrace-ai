import { Injectable, Logger } from '@nestjs/common';

import { EmbeddingService } from './embedding.service';
import { ProductClientService } from '../../integrations/go-core/product-client.service';

@Injectable()
export class ReindexService {
  private readonly logger = new Logger(ReindexService.name);

  constructor(
    private readonly embeddingService: EmbeddingService,
    private readonly productClient: ProductClientService,
  ) { }

  async reindexAll(): Promise<void> {
    let page = 1;
    const limit = 100;

    while (true) {
      const result =
        await this.productClient.getProducts(page, limit);

      if (result.data.length === 0) {
        break;
      }

      for (const product of result.data) {
        await this.embeddingService.processEvent({
          eventId: `reindex-${product.id}`,
          eventType: 'product.reindexed',
          eventVersion: '1.0',
          producer: 'nest-ai-service',
          correlationId: `reindex-${product.id}`,
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

    this.logger.log('Reindex completed');
  }
}