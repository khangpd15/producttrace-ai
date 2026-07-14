import { HttpException, HttpStatus, Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import * as fs from 'fs';
import { v5 as uuidv5 } from 'uuid';

import { BuildTextUtil } from '../../utils/build-text.util';
import { Event } from '../../messaging/types/event.interface';
import { RabbitMQProducerService } from '../../messaging/rabbitmq/rabbitmq-producer.service';
import { RABBITMQ } from '../../messaging/rabbitmq/rabbitmq.constants';

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
    private readonly rabbitmqProducer: RabbitMQProducerService,
  ) {
    const isDocker = fs.existsSync('/.dockerenv');
    const defaultEmbeddingUrl = isDocker ? 'http://embedding-service:8000' : 'http://localhost:8000';
    const envEmbeddingUrl = this.configService.get('EMBEDDING_SERVICE_URL');
    const resolvedEmbeddingUrl = !isDocker && envEmbeddingUrl?.includes('embedding-service')
      ? defaultEmbeddingUrl
      : envEmbeddingUrl ?? defaultEmbeddingUrl;
    this.apiUrl = resolvedEmbeddingUrl;
  }

  async processEvent(event: Event): Promise<void> {
    if (this.processedEvents.has(event.event_id)) {
      this.logger.warn(`[EMBED][SKIP_DUPLICATE] ${event.event_id}`);
      return;
    }

    this.processedEvents.add(event.event_id);

    if (this.processedEvents.size > this.MAX_CACHE) {
      this.processedEvents.clear();
    }

    try {
      // 1. RECEIVED EVENT
      this.logger.log(`[EMBED][RECEIVED]`, {
        eventType: event.event_type,
        eventId: event.event_id,
      });

      const text = BuildTextUtil.buildText(event);

      if (!text || text.trim().length === 0) {
        throw new Error('Empty embedding text');
      }

      // 2. TEXT GENERATED
      this.logger.log(`[EMBED][TEXT]`, {
        length: text.length,
        preview: text.slice(0, 200),
      });

      const vector = await this.generateEmbedding(text);

      // 3. VECTOR GENERATED
      this.logger.log(`[EMBED][VECTOR]`, {
        dim: vector.length,
      });

      // 4. SAFE DETERMINISTIC ID
      const pointId = this.buildPointId(event);

      const payload = this.buildPayload(event);

      const embeddingGeneratedEvent = {
        eventId: event.event_id,
        productId: this.extractProductId(event),
        pointId,
        vector,
        payload,
        timestamp: new Date().toISOString(),
      };

      this.logger.log(`[EMBED][PIPELINE]`, {
        step: 'PUBLISH_EMBEDDING_GENERATED',
        eventId: event.event_id,
        pointId,
      });

      // 5. PUBLISH EMBEDDING GENERATED EVENT
      await this.rabbitmqProducer.emit(
        RABBITMQ.ROUTING_KEYS.EMBEDDING_GENERATED,
        embeddingGeneratedEvent,
      );

      this.logger.log(`[EMBED][EMBEDDING_GENERATED]`, {
        eventId: event.event_id,
        pointId,
      });

    } catch (error) {
      this.logger.error(
        `Failed processing event ${event.event_id}: ${error instanceof Error
          ? error.message
          : JSON.stringify(error)
        }`,
      );

      throw error;
    }
  }

  private buildPointId(event: Event): string {
    const productId = this.extractProductId(event);

    if (!productId || productId === 'unknown') {
      this.logger.error(`[EMBED][INVALID_PRODUCT_ID]`, event);
      throw new Error('Missing productId for embedding');
    }

    return uuidv5(productId, this.NAMESPACE);
  }

  async generateEmbedding(text: string): Promise<number[]> {
    const response = await fetch(`${this.apiUrl}/embed`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ text }),
    });

    if (!response.ok) {
      const body = await response.text();

      this.logger.error(
        `Embedding API failed: ${response.status} ${body}`,
      );

      throw new HttpException(
        'Embedding service error',
        HttpStatus.BAD_GATEWAY,
      );
    }

    const data =
      (await response.json()) as EmbeddingResponse;

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
      eventType: event.event_type,
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
      this.logger.warn(
        `[EMBED][MISSING_PRODUCT_ID]`,
        event,
      );

      return 'unknown';
    }

    return productId;
  }
}