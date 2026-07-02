import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';

import { SyncConsumer } from './sync.consumer';
import { SyncService } from './sync.service';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';
import { KafkaModule } from '../../kafka/kafka.module';

@Module({
  imports: [ConfigModule, KafkaModule],
  controllers: [SyncConsumer],
  providers: [
    SyncService,
    QdrantService,
  ],
})
export class SyncModule {}