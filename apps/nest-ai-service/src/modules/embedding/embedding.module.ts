import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';

import { EmbeddingConsumer } from '../../kafka/embedding.consumer';
import { EmbeddingService } from './embedding.service';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';

@Module({
  imports: [ConfigModule],
  controllers: [EmbeddingConsumer],
  providers: [EmbeddingService, QdrantService],
})
export class EmbeddingModule {}
