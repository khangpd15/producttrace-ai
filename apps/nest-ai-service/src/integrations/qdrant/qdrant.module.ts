import { Module } from '@nestjs/common';
import { QdrantController } from './qdrant.controller'; 
import { QdrantService } from './qdrant.service'; 

@Module({
    imports: [],
  controllers: [QdrantController], 
  providers: [QdrantService],
  exports: [QdrantService],
})
export class QdrantModule {}