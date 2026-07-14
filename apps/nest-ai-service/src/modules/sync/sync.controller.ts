import { Controller } from '@nestjs/common';
import { EventPattern, Payload } from '@nestjs/microservices';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';

@Controller()
export class SyncController {
    constructor(private readonly qdrantService: QdrantService) { }

    // Lắng nghe sự kiện từ bên Go truyền qua RabbitMQ
    @EventPattern('product.created')
    async handleProductCreated(@Payload() data: { id: string; name: string; description: string }) {
        console.log('─── Nhận dữ liệu sync từ Go ───', data);

        // Tạo vector dummy (chức năng AI  sẽ nằm ở thư mục modules/embedding)
        const dummyVector = Array.from({ length: 1536 }, () => Math.random());

        // Gọi hàm upsert 
        await this.qdrantService.upsertProduct({
            id: data.id,
            vector: dummyVector,
            name: data.name,
            metadata: {
                name: data.name,
                description: data.description
            },
        });

        console.log(`─── Has been [${data.name}] into Qdrant DB successfully! ───`);
    }
}