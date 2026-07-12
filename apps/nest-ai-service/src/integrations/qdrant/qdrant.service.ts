import { Injectable, OnModuleInit } from '@nestjs/common';
import { ConfigService } from '@nestjs/config/dist/config.service';
import { QdrantClient } from '@qdrant/js-client-rest';

@Injectable()
export class QdrantService implements OnModuleInit {
  private client: QdrantClient;
  private collectionName = 'producttrace_geo_collection';

  constructor(private configService: ConfigService) {
    const qdrantUrl = this.configService.get<string>('QDRANT_URL') || 'http://localhost:6333';
    this.collectionName = 'producttrace_geo_collection'; 
    this.client = new QdrantClient({ url: qdrantUrl });
  }

  async onModuleInit() {
    try {
      const collections = await this.client.getCollections();
      const hasCollection = collections.collections.some(c => c.name === this.collectionName);

      if (!hasCollection) {
        await this.client.createCollection(this.collectionName, {
          vectors: {}, 
        });
        
        await this.client.createPayloadIndex(this.collectionName, {
          field_name: 'location',
          field_schema: 'geo',
          wait: true,
        });
        console.log(`[QdrantService] Geo collection initialized with geo index.`);
      }
    } catch (error) {
      console.error('[QdrantService] Failed to initialize Qdrant collection:', error);
    }
  }

  async findStoresByRadius(lat: number, lng: number, radiusMeters: number) {
    return this.findLocationsByRadius(lat, lng, radiusMeters, 'store');
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
  // Hàm bọc lưu/cập nhật thông tin Store 
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
        vector: {},
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
        vector: {},
      }],
    });
  }
}