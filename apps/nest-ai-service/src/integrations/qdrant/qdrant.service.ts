import { Injectable, OnModuleInit } from '@nestjs/common';
import { ConfigService } from '@nestjs/config/dist/config.service';
import { QdrantClient } from '@qdrant/js-client-rest';

@Injectable()
export class QdrantService implements OnModuleInit {
  private client: QdrantClient;
  private readonly collectionName = 'product_embeddings';
  private readonly vectorSize = 1024; // Cập nhật theo model BGE-M3
  private collectionReady = false;
  private ensureCollectionPromise: Promise<void> | null = null;

  constructor(private configService: ConfigService) {
    const qdrantUrl = this.configService.get<string>('QDRANT_URL') || 'http://localhost:6333';
    this.collectionName = 'producttrace_geo_collection'; 
    this.client = new QdrantClient({ url: qdrantUrl });
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

      if (!hasCollection) {
        await this.client.createCollection(this.collectionName, {
          vectors: {}, 
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

  // Tìm địa điểm theo vị trí địa lý và loại thực thể
  async findLocationsByRadius(
    lat: number,
    lng: number,
    radiusMeters: number,
    type: 'store' | 'service_center',
  ) {
    try {
      const result = await this.client.scroll(this.collectionName, {
        filter: {
          must: [
            { key: 'type', match: { value: type } }, 
            {
              key: 'location',
              geo_radius: {
                center: { lat, lon: lng }, 
                radius: radiusMeters,
              },
            },
          ],
        },
        with_payload: true,
      });
      return result.points; 
    } catch (error) {
      console.error(`[QdrantService] Error in findLocationsByRadius for ${type}:`, error);
      throw error;
    }
  }

  // Tìm địa điểm chứa sản phẩm cụ thể
  async findProductsByRadius(lat: number, lng: number, radiusMeters: number, productId?: string) {
    try {
      const filters: any[] = [
        {
          key: 'location',
          geo_radius: {
            center: { lat, lon: lng }, 
            radius: radiusMeters,
          },
        },
      ];

      if (productId) {
        filters.push({ key: 'products', match: { value: productId } });
      }

      const result = await this.client.scroll(this.collectionName, {
        filter: { must: filters },
        with_payload: true,
      });

      return result.points;
    } catch (error) {
      console.error('[QdrantService] Error in findProductsByRadius:', error);
      throw error;
    }
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
          type: data.type || 'store',
          location: { lat: data.latitude, lon: data.longitude },
          address: data.address,
          products: data.products || [],
        },
        vector: Array(this.vectorSize).fill(0), // Cần khởi tạo mảng vector rỗng đúng size
      }],
    });
  }

  // Hàm bọc lưu/cập nhật thông tin Product 
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