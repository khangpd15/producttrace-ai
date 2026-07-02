import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';

import { EmbeddingModule } from './modules/embedding/embedding.module';

import { ProductCreatedConsumer } from './messaging/consumers/product-created.consumer';
import { UserRegisteredConsumer } from './messaging/consumers/user-registered.consumer';
import { PasswordResetConsumer } from './messaging/consumers/password-reset.consumer';

import { MailModule } from './modules/mail/mail.module';
import { MockController } from './mock.controller';
import { AuthModule } from './auth/auth.module';
import { SyncModule } from './modules/sync/sync.module';
import { KafkaModule } from './kafka/kafka.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: '.env',
    }),
    KafkaModule,
    EmbeddingModule,
    SyncModule,
    MailModule,
    AuthModule,
  ],
  controllers: [
    MockController,
  ],
  providers: [
    ProductCreatedConsumer,
    UserRegisteredConsumer,
    PasswordResetConsumer,
  ],
})
export class AppModule { }