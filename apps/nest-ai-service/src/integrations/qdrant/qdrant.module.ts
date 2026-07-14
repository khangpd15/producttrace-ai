import { Module } from '@nestjs/common';
import { QdrantService } from './qdrant.service';
import { QdrantController } from './qdrant.controller';

@Module({
  controllers: [QdrantController],
  providers: [QdrantService],
  exports: [QdrantService],
})
export class QdrantModule {}