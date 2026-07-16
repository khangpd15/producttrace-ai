import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';

import { SyncModule } from './modules/sync/sync.module';
import { EmbeddingModule } from './modules/embedding/embedding.module';
import { ReindexModule } from './modules/reindex/reindex.module';

import { MailModule } from './modules/mail/mail.module';
import { MockController } from './mock.controller';
import { AuthModule } from './auth/auth.module';

// RabbitMQ của Email
import { RabbitMQModule } from './messaging/rabbitmq/rabbitmq.module';

@Module({
  imports: [
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
  ],
  controllers: [MockController],
})
export class AppModule {}