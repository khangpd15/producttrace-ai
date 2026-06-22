import { HttpException, HttpStatus, Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { v5 as uuidv5 } from 'uuid';

import { BuildTextUtil } from '../../utils/build-text.util';
import { Event } from '../../messaging/types/event.interface';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';

interface EmbeddingResponse {
  vector: number[];
}

@Injectable()
export class EmbeddingService {
  private readonly logger = new Logger(EmbeddingService.name);
  private readonly apiUrl: string;
  private processedEvents = new Set<string>();
  private readonly MAX_CACHE = 10000;

  private readonly NAMESPACE =
    '6ba7b810-9dad-11d1-80b4-00c04fd430c8';

  constructor(
    private readonly configService: ConfigService,
    private readonly qdrantService: QdrantService,
  ) {
    this.apiUrl =
      this.configService.get('EMBEDDING_SERVICE_URL') ??
      'http://embedding-service:8000';
  }

  async processEvent(event: Event): Promise<void> {

    if (this.processedEvents.has(event.eventId)) {
      this.logger.warn(`[EMBED][SKIP_DUPLICATE] ${event.eventId}`);
      return;
    }

    this.processedEvents.add(event.eventId);

    if (this.processedEvents.size > this.MAX_CACHE) {
      this.processedEvents.clear();
    }

    // 1. RECEIVED EVENT
    this.logger.log(`[EMBED][RECEIVED]`, {
      eventType: event.eventType,
      eventId: event.eventId,
    });

    const text = BuildTextUtil.buildText(event);

    if (!text || text.trim().length === 0) {
      throw new Error('Empty embedding text');
    }

    // 2. TEXT GENERATED
    this.logger.log(`[EMBED][TEXT]`, {
      length: text?.length,
      preview: text?.slice(0, 200),
    });

    const vector = await this.generateEmbedding(text);

    // 3. VECTOR GENERATED
    this.logger.log(`[EMBED][VECTOR]`, {
      dim: vector?.length,
    });

    // 4. SAFE DETEMINISTIC ID (FIXED)
    const pointId = this.buildPointId(event);

    const payload = this.buildPayload(event);

    this.logger.log(`[EMBED][PIPELINE]`, {
      step: 'QDRANT_UPSERT',
      eventId: event.eventId,
      pointId,
    });

    // 5. UPSERT TO QDRANT
    await this.qdrantService.upsertVector(pointId, vector, payload);

    this.logger.log(`[EMBED][UPSERT_DONE]`, {
      pointId,
      eventId: event.eventId,
    });

    this.logger.log(`[EMBED][QDRANT_SYNCED]`, {
      eventId: event.eventId,
      pointId,
    });
  }

  private buildPointId(event: Event): string {
    const productId = this.extractProductId(event);

    if (!productId || productId === 'unknown') {
      this.logger.error(`[EMBED][INVALID_PRODUCT_ID]`, event);
      throw new Error('Missing productId for embedding');
    }

    return uuidv5(productId, this.NAMESPACE);
  }

  private async generateEmbedding(text: string): Promise<number[]> {
    const response = await fetch(`${this.apiUrl}/embed`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text }),
    });

    if (!response.ok) {
      const body = await response.text();
      this.logger.error(`Embedding API failed: ${response.status} ${body}`);
      throw new HttpException(
        'Embedding service error',
        HttpStatus.BAD_GATEWAY,
      );
    }

    const data = (await response.json()) as EmbeddingResponse;

    if (!Array.isArray(data.vector)) {
      this.logger.error('Invalid embedding response');
      throw new HttpException(
        'Invalid embedding response',
        HttpStatus.BAD_GATEWAY,
      );
    }

    return data.vector;
  }

  private buildPayload(event: Event): Record<string, unknown> {
    return {
      productId: this.extractProductId(event),
      eventType: event.eventType,
      timestamp: event.timestamp,
    };
  }

  private extractProductId(event: Event): string {
    const payload = event.payload as Record<string, unknown>;

    const productId = String(
      payload?.productId ??
      payload?.product ??
      payload?.productID ??
      payload?.id ??
      '',
    );

    if (!productId) {
      this.logger.warn(`[EMBED][MISSING_PRODUCT_ID]`, event);
      return 'unknown';
    }

    return productId;
  }
}