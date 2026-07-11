import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';

import { SyncConsumer } from './sync.consumer';
import { SyncService } from './sync.service';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';
import { EmbeddingRabbitMQModule } from '../../integrations/rabbitmq/ai-rabbitmq.module';


@Module({
  imports: [ConfigModule, EmbeddingRabbitMQModule],
  controllers: [SyncConsumer],
  providers: [
    SyncService,
    QdrantService,
  ],
})
export class SyncModule {}