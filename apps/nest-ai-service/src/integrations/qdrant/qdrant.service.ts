import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { QdrantClient } from '@qdrant/qdrant-js';
import { SearchResult } from '../../modules/search/interfaces/search-result.interface';

@Injectable()
export class QdrantService implements OnModuleInit {
  private readonly logger = new Logger(QdrantService.name);
  private client: QdrantClient;
  private readonly collectionName = 'product_embeddings';
  private readonly vectorSize = 1024; // Cập nhật theo model BGE-M3
  private collectionReady = false;
  private ensureCollectionPromise: Promise<void> | null = null;

  constructor(private readonly configService: ConfigService) {
    this.client = new QdrantClient({
      url: this.configService.get('QDRANT_URL'),
      apiKey: this.configService.get('QDRANT_API_KEY'),
    });
  }

  async onModuleInit() {
    await this.ensureCollection();
  }

  // =========================
  // SINGLETON ENSURE COLLECTION
  // =========================
  private async ensureCollection(): Promise<void> {
    if (this.collectionReady) return;
    if (!this.ensureCollectionPromise) {
      this.ensureCollectionPromise = this.doEnsureCollection().finally(() => {
        this.ensureCollectionPromise = null;
      });
    }
    return this.ensureCollectionPromise;
  }

  private async doEnsureCollection(): Promise<void> {
    try {
      const collections = await this.client.getCollections();
      const exists = collections.collections.some((c) => c.name === this.collectionName);

      if (!exists) {
        this.logger.log(`Initializing collection: ${this.collectionName}`);
        await this.client.createCollection(this.collectionName, {
          vectors: { size: this.vectorSize, distance: 'Cosine' },
        });
        await this.client.createPayloadIndex(this.collectionName, { field_name: 'location', field_schema: 'geo' });
        await this.client.createPayloadIndex(this.collectionName, { field_name: 'type', field_schema: 'keyword' });
      }
      this.collectionReady = true;
    } catch (error) {
      this.logger.error('Failed to initialize collection:', error);
      throw error;
    }
  }

  // =========================
  // VECTOR SEARCH & UPSERT
  // =========================
  async upsertVector(id: string, vector: number[], payload: Record<string, unknown>): Promise<void> {
    await this.ensureCollection();
    await this.client.upsert(this.collectionName, {
      wait: true,
      points: [{ id, vector, payload }],
    });
  }

  async vectorSearch(vector: number[], filter?: Record<string, unknown>, limit = 10): Promise<SearchResult[]> {
    await this.ensureCollection();
    const result = await this.client.search(this.collectionName, {
      vector,
      filter: filter as any,
      limit,
      with_payload: true,
    });
    return result as unknown as SearchResult[];
  }

  // =========================
  // GEO & PRODUCT SEARCH
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
  // UPSERT ENTITIES
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
        vector: Array(this.vectorSize).fill(0), // Cần khởi tạo mảng vector rỗng đúng size
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
        vector: Array(this.vectorSize).fill(0),
      }],
    });
  }
}