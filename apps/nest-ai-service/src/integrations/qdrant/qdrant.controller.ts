import { Controller, Post, Body } from '@nestjs/common';
import { QdrantService } from './qdrant.service';

@Controller('qdrant-test')
export class QdrantController {
  constructor(private readonly qdrantService: QdrantService) { }

  @Post('upsert-dummy')
  async upsertDummy() {
    // Tạo 1 vector mẫu 768 chiều
    const dummyVector = new Array(768).fill(0.1);
    const dummyId = 'product_test_001';
    const dummyMetadata = { name: 'Dummy Productte', price: 100000 };

    await this.qdrantService.upsertProduct({
      id: dummyId,
      vector: dummyVector,
      metadata: dummyMetadata
    });
    return { message: 'Successfully upserted dummy product!' };
  }
}