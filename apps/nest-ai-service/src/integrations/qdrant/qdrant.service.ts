import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { QdrantClient } from '@qdrant/qdrant-js';
import { SearchResult } from '../../modules/search/interfaces/search-result.interface';

@Injectable()
export class QdrantService implements OnModuleInit {
  private readonly logger = new Logger(QdrantService.name);
  private client: QdrantClient;
  private readonly collectionName = 'product_embeddings';
<<<<<<< HEAD
  private readonly vectorSize = 768; // ✅ Cố định size vector theo model
=======

  // ✅ FIXED VECTOR SIZE (quan trọng cho production)
  private readonly vectorSize = 1024; // đổi theo model BGE-M3

  private collectionReady = false;
  // Singleton Promise: concurrent callers all await the same creation task
  // instead of returning early (the previous bool-flag race condition).
  private ensureCollectionPromise: Promise<void> | null = null;
>>>>>>> 894a522de7509c50f12999bf98d08c64e9fe1486

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
<<<<<<< HEAD
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
=======
  async upsertVector(
    id: string,
    vector: number[],
    payload: Record<string, unknown>,
  ): Promise<void> {
    // validate vector
    if (!vector || vector.length !== this.vectorSize) {
      throw new Error(
        `Invalid vector size: expected ${this.vectorSize}, got ${vector?.length}`,
      );
    }

    // ensure collection exists
    await this.ensureCollection();

    const response = await fetch(
      `${this.baseUrl}/collections/${this.collectionName}/points?wait=true`,
      {
        method: 'PUT',
        headers: this.headers,
        body: JSON.stringify({
          points: [
            {
              id,
              vector,
              payload,
            },
          ],
        }),
      },
    );

    if (!response.ok) {
      const body = await response.text();
      this.logger.error(`Qdrant upsert failed: ${response.status} ${body}`);
      throw new Error(`Qdrant upsert failed: ${response.status}`);
    }

    this.logger.log(
      `Upsert success: id=${id}, collection=${this.collectionName}`,
    );
  }

  // =========================
  // VECTOR SEARCH
  // =========================
  async vectorSearch(
    vector: number[],
    filter?: Record<string, unknown>,
    limit = 10,
  ): Promise<SearchResult[]> {
    // validate vector
    if (!vector || vector.length !== this.vectorSize) {
      throw new Error(
        `Invalid vector size: expected ${this.vectorSize}, got ${vector?.length}`,
      );
    }

    // ensure collection exists
    await this.ensureCollection();

    const response = await fetch(
      `${this.baseUrl}/collections/${this.collectionName}/points/search`,
      {
        method: 'POST',
        headers: this.headers,
        body: JSON.stringify({
          vector,
          filter,
          limit,
          with_payload: true,
          with_vector: false,
        }),
      },
    );

    if (!response.ok) {
      const body = await response.text();

      this.logger.error(
        `Qdrant search failed: ${response.status} ${body}`,
      );

      throw new Error(
        `Qdrant search failed: ${response.status}`,
      );
    }

    const body = await response.json();

    return body.result as SearchResult [];
  }
  // =========================
  // ENSURE COLLECTION
  // =========================
  private async ensureCollection(): Promise<void> {
    // Fast path — collection already confirmed ready
    if (this.collectionReady) return;

    // Singleton Promise pattern: if creation is already in progress, all
    // concurrent callers await the same Promise instead of returning early
    // (the previous bool-flag approach caused a race condition where the
    // second caller skipped creation and immediately tried to upsert into a
    // non-existent collection).
    if (!this.ensureCollectionPromise) {
      this.ensureCollectionPromise = this.doEnsureCollection().finally(() => {
        this.ensureCollectionPromise = null;
      });
    }

    return this.ensureCollectionPromise;
  }

  private async doEnsureCollection(): Promise<void> {
    const exists = await this.collectionExists();

    if (!exists) {
      this.logger.log(
        `Creating collection=${this.collectionName}, vectorSize=${this.vectorSize}`,
      );

      const response = await fetch(
        `${this.baseUrl}/collections/${this.collectionName}`,
        {
          method: 'PUT',
          headers: this.headers,
          body: JSON.stringify({
            vectors: {
              size: this.vectorSize,
              distance: 'Cosine',
            },
          }),
        },
      );

      if (!response.ok) {
        const body = await response.text();
        throw new Error(
          `Qdrant collection creation failed: ${response.status} ${body}`,
        );
      }

      this.logger.log(`Collection created: ${this.collectionName}`);
    } else {
      await this.validateCollectionConfig();
>>>>>>> 894a522de7509c50f12999bf98d08c64e9fe1486
    }

    this.collectionReady = true;
  }

  private async validateCollectionConfig(): Promise<void> {
    const response = await fetch(
      `${this.baseUrl}/collections/${this.collectionName}`,
      {
        method: 'GET',
        headers: this.headers,
      },
    );

    if (!response.ok) {
      const body = await response.text();
      throw new Error(
        `Failed to validate Qdrant collection: ${response.status} ${body}`,
      );
    }

    const info = await response.json();
    // Qdrant v1.x response: result.config.params.vectors.{size, distance}
    const vectorsCfg =
      info?.result?.vectors ??
      info?.result?.config?.params?.vectors;

    const currentSize: number | undefined =
      vectorsCfg?.size ?? vectorsCfg?.vector_size;

    // Normalise to lowercase for safe comparison ('Cosine' vs 'cosine')
    const currentDistance: string = String(
      vectorsCfg?.distance ?? vectorsCfg?.distance_metric ?? '',
    ).toLowerCase();

    if (currentSize !== this.vectorSize || currentDistance !== 'cosine') {
      throw new Error(
        `Qdrant collection ${this.collectionName} config mismatch: ` +
        `expected size=${this.vectorSize}, distance=Cosine but found ` +
        `size=${currentSize}, distance=${currentDistance}. ` +
        `Delete and recreate the collection before upserting vectors.`,
      );
    }

    this.logger.log(
      `[QDRANT] Collection config validated: size=${currentSize}, distance=${currentDistance}`,
    );
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