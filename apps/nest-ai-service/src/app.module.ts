import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';

import { EmbeddingModule } from './modules/embedding/embedding.module';
import { SyncModule } from './modules/sync/sync.module';

import { MailModule } from './modules/mail/mail.module';
import { MockController } from './mock.controller';
import { AuthModule } from './auth/auth.module';

// RabbitMQ của Email
import { RabbitMQModule } from './messaging/rabbitmq/rabbitmq.module';
import { EmbeddingRabbitMQModule } from './integrations/rabbitmq/ai-rabbitmq.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: '.env',
    }),
    EmbeddingModule,
    SyncModule,
    MailModule,
    AuthModule,
    RabbitMQModule,
    EmbeddingRabbitMQModule,
  ],
  controllers: [MockController],
})
export class AppModule {}