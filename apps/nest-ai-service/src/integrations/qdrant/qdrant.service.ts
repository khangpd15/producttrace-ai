import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { QdrantClient } from '@qdrant/qdrant-js';
import { SearchResult } from '../../modules/search/interfaces/search-result.interface';

@Injectable()
export class QdrantService implements OnModuleInit {
  private readonly logger = new Logger(QdrantService.name);
  private client: QdrantClient;
  private readonly collectionName = 'product_embeddings';
  private readonly vectorSize = 768; // ✅ Cố định size vector theo model

  constructor(private readonly configService: ConfigService) {
    this.client = new QdrantClient({
      url: this.configService.get('QDRANT_URL'),
      apiKey: this.configService.get('QDRANT_API_KEY'),
    });
  }

  async onModuleInit() {
    await this.initializeCollection();
  }

  // =========================
  // KHỞI TẠO COLLECTION (Clean)
  // =========================
  private async initializeCollection(): Promise<void> {
    try {
      const collections = await this.client.getCollections();
      const exists = collections.collections.some((c: { name: string; }) => c.name === this.collectionName);

      if (!exists) {
        this.logger.log(`Initializing collection: ${this.collectionName}`);
        await this.client.createCollection(this.collectionName, {
          vectors: { size: this.vectorSize, distance: 'Cosine' },
        });

        // Index cho Geo và Type
        await this.client.createPayloadIndex(this.collectionName, { field_name: 'location', field_schema: 'geo' });
        await this.client.createPayloadIndex(this.collectionName, { field_name: 'type', field_schema: 'keyword' });
        
        console.log(`[QdrantService] Collection initialized successfully.`);
      }
    } catch (error) {
      this.logger.error('[QdrantService] Failed to initialize collection:', error);
    }
  }

  // =========================
  // VECTOR SEARCH & UPSERT
  // =========================
  async upsertVector(id: string, vector: number[], payload: Record<string, unknown>): Promise<void> {
    await this.client.upsert(this.collectionName, {
      wait: true,
      points: [{ id, vector, payload }],
    });
  }

  async vectorSearch(vector: number[], filter?: Record<string, unknown>, limit = 10): Promise<SearchResult[]> {
    const result = await this.client.search(this.collectionName, {
      vector,
      filter: filter as any,
      limit,
      with_payload: true,
    });
    return result as unknown as SearchResult[];
  }

  // =========================
  // GEO & PRODUCT SEARCH (Giữ lại logic của bạn)
  // =========================
  async findStoresByRadius(lat: number, lon: number, radiusMeters: number) {
    return this.findLocationsByRadius(lat, lon, radiusMeters, 'store');
  }

  async findLocationsByRadius(lat: number, lon: number, radiusMeters: number, type: 'store' | 'service_center') {
    const result = await this.client.scroll(this.collectionName, {
      filter: {
        must: [
          { key: 'type', match: { value: type } },
          { key: 'location', geo_radius: { center: { lat, lon }, radius: radiusMeters } },
        ],
      },
      with_payload: true,
    });
    return result.points;
  }

  async findProductsByRadius(lat: number, lon: number, radiusMeters: number, productId?: string) {
    const filters: any[] = [
      { key: 'location', geo_radius: { center: { lat, lon }, radius: radiusMeters } },
    ];
    if (productId) filters.push({ key: 'products', match: { value: productId } });

    const result = await this.client.scroll(this.collectionName, {
      filter: { must: filters },
      with_payload: true,
    });
    return result.points;
  }

  // =========================
  // UPSERT ENTITIES (Giữ lại logic của bạn)
  // =========================
  async upsertStoreToQdrant(data: any) {
    return this.client.upsert(this.collectionName, {
      wait: true,
      points: [{
        id: data.id,
        payload: {
          name: data.name,
          type: 'store',
          location: { lat: data.latitude, lon: data.longitude },
          address: data.address,
          products: data.products || [],
        },
        vector: [],
      }],
    });
  }

  async upsertProduct(data: any) {
    return this.client.upsert(this.collectionName, {
      wait: true,
      points: [{
        id: data.id,
        payload: {
          name: data.name,
          type: 'product',
          productId: data.productId,
          metadata: data.metadata || {},
        },
        vector: [],
      }],
    });
  }
}