import { Injectable, Logger } from '@nestjs/common';
import { EmbeddingService } from '../embedding/embedding.service';
import { ProductClientService, ProductDetail } from '../../integrations/go-core/product-client.service';

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

      for (const item of result.data) {
        // Fetch full product detail to get variants, tags, metadata, etc.
        const product = await this.productClient.getProductById(item.id);
        if (!product) {
          this.logger.warn(`[REINDEX][SKIP] Product ${item.id} not found`);
          continue;
        }

        const runId = Math.floor(Date.now() / 1000);

        // Build a payload that matches the product.created event structure
        const payload = this.buildProductPayload(product);

        await this.embeddingService.processEvent({
          event_id: `reindex-${product.id}-${runId}`,
          event_type: 'product.reindexed',
          event_version: '1.0',
          producer: 'nest-ai-service',
          correlation_id: `reindex-${product.id}-${runId}`,
          timestamp: new Date().toISOString(),
          payload,
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

  /**
   * Builds a payload matching the product.created event structure.
   * This ensures reindexed embeddings are identical to real-time embeddings.
   */
  private buildProductPayload(product: ProductDetail): Record<string, unknown> {
    const payload: Record<string, unknown> = {
      id: product.id,
      productId: product.id,
      name: product.name,
      description: product.description ?? '',
      slug: product.slug ?? '',
      status: product.status ?? '',
      thumbnail_url: product.thumbnail_url ?? '',
      tags: product.tags ?? [],
      metadata: product.metadata ?? {},
      created_at: product.created_at ?? '',
      updated_at: product.updated_at ?? '',
      created_by: product.created_by ?? '',
    };

    // Map variants to match the product.created event structure
    if (product.variants && product.variants.length > 0) {
      payload.variants = product.variants.map(v => ({
        id: v.id,
        sku: v.sku,
        name: v.name,
        barcode: v.barcode ?? null,
        price: v.price ?? null,
        currency: v.currency ?? null,
        images: v.images ?? [],
        status: v.status ?? null,
        attributes: (v.attributes ?? []).map(a => ({
          id: a.id,
          attribute_id: a.attribute_id,
          label: a.label,
          value_text: a.value_text ?? null,
          value_number: a.value_number ?? null,
          value_boolean: a.value_boolean ?? null,
        })),
      }));
    } else {
      payload.variants = [];
    }

    return payload;
  }
}