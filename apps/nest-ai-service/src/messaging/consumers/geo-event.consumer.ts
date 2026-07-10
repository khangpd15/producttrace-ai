import { Controller } from '@nestjs/common';
import { EventPattern, Payload } from '@nestjs/microservices';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';

@Controller()
export class GeoEventConsumer {
  constructor(private readonly qdrantService: QdrantService) {}

  // Lắng nghe sự kiện từ bên Go bắn sang RabbitMQ khi có cửa hàng/trạm bảo hành mới hoặc cập nhật
  @EventPattern('store.created')
  async handleStoreCreated(@Payload() data: any) {
    console.log('[RabbitMQ Consumer] Received store.created event from Go:', data);
    
    try {
      // Ép kiểu client và collectionName từ QdrantService để tận dụng kết nối có sẵn
      const client = (this.qdrantService as any).client;
      const collectionName = (this.qdrantService as any).collectionName;

      // Tự động bốc tách dữ liệu từ Go và cấu trúc lại để lưu xuống Qdrant Geo Index
      await client.upsert(collectionName, {
        wait: true,
        points: [{
          id: data.id,
          payload: {
            name: data.name,
            type: data.type || 'store',        // Phân loại 'store' hoặc 'service_center'
            location: { 
              lat: data.latitude, 
              lon: data.longitude 
            },
            address: data.address,
            products: data.products || [],     // Mảng danh sách sản phẩm bán tại đó (nếu có)
          },
          vector: {}, 
        }],
      });

      console.log(`[Qdrant] Automatically synchronized location: ${data.name}`);
    } catch (error) {
      console.error('[RabbitMQ Consumer] Failed to sync data to Qdrant:', error);
    }
  }
}