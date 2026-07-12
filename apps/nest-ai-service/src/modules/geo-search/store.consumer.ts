import { Controller } from '@nestjs/common';
import { EventPattern, Payload } from '@nestjs/microservices';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';

// Định nghĩa khuôn dữ liệu mà bên Go sẽ bắn sang
interface StoreCreatedPayload {
  id: number;
  name: string;
  address: string;
  latitude: number;
  longitude: number;
}

@Controller()
export class StoreConsumer {
  constructor(private readonly qdrantService: QdrantService) {}

  // 1. Lắng nghe event khi bên Go tạo cửa hàng thành công
  @EventPattern('store.created')
  async handleStoreCreated(@Payload() data: StoreCreatedPayload) {
    console.log('📥 [NestJS Consumer] Received store.created event from Go:', data);

    try {
      // Gọi Qdrant Client để cập nhật vùng không gian
      await this.qdrantService.upsertStoreToQdrant({
        id: data.id,
        name: data.name,
        lat: data.latitude,
        lng: data.longitude,
        address: data.address
      });
      console.log(`✅ Synchronized store ID ${data.id} to Qdrant successfully!`);
    } catch (error) {
      console.error(`❌ Failed to sync store ID ${data.id} to Qdrant:`, error);
    }
  }
}