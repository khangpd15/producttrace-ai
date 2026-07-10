import { Injectable, OnModuleInit } from '@nestjs/common';
import { QdrantClient } from '@qdrant/js-client-rest';

@Injectable()
export class QdrantService implements OnModuleInit {
  private client: QdrantClient;
  private collectionName = 'producttrace_geo_collection';

  constructor() {
    // Kết nối tới Qdrant đang chạy trong Docker
    this.client = new QdrantClient({ url: 'http://localhost:6333' });
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
}