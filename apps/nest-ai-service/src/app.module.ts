import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';

import { SyncModule } from './modules/sync/sync.module';
import { EmbeddingModule } from './modules/embedding/embedding.module';
import { ReindexModule } from './modules/reindex/reindex.module';

import { MailModule } from './modules/mail/mail.module';
import { MockController } from './mock.controller';
import { AuthModule } from './auth/auth.module';
import { GeoEventConsumer } from './messaging/consumers/geo-event.consumer';
import { QdrantModule } from './integrations/qdrant/qdrant.module';
import { GeoSearchModule } from './modules/geo-search/geo-search.module';

// RabbitMQ của Email
import { RabbitMQModule } from './messaging/rabbitmq/rabbitmq.module';
import { GeoSearchModule } from './modules/geo-search/geo-search.module';
import { QdrantModule } from './integrations/qdrant/qdrant.module';

@Module({
  imports: [
    GeoSearchModule,
    QdrantModule,
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: '.env',
    }),
    SyncModule,
    EmbeddingModule,
    ReindexModule,
    MailModule,
    AuthModule,
    RabbitMQModule,
    GeoSearchModule, 
    QdrantModule,
  ],
  controllers: [MockController],
})
export class AppModule {}