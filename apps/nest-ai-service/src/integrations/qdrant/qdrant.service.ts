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
        // Khởi tạo collection lưu trữ hình học
        await this.client.createCollection(this.collectionName, {
          vectors: {}, // Sử dụng cấu trúc lưu trữ không gian (Non-vector)
        });
        
        // Kích hoạt Geo Index tăng tốc quét bán kính
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

  // CƠ CHẾ CORE DÙNG CHUNG (Lọc theo vị trí địa lý và loại thực thể)
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
            { key: 'type', match: { value: type } }, // Lọc store hoặc service_center
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

      // Map lại cấu trúc dữ liệu trả về mảng gọn gàng
      return result.points.map(p => ({
        id: p.id,
        ...(p.payload as object),
      }));
    } catch (error) {
      console.error(`[QdrantService] Error in findLocationsByRadius for ${type}:`, error);
      throw error;
    }
  }

  // Tìm địa điểm chứa sản phẩm cụ thể trong bán kính (km)
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

      // Nếu người dùng truyền mã sản phẩm, dùng bộ lọc match của Qdrant để quét trong mảng 'products'
      if (productId) {
        filters.push({ key: 'products', match: { value: productId } });
      }

      const result = await this.client.scroll(this.collectionName, {
        filter: { must: filters },
        with_payload: true,
      });

      return result.points.map(p => ({
        id: p.id,
        ...(p.payload as object),
      }));
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