import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';

import { MailModule } from './modules/mail/mail.module';
import { MockController } from './mock.controller';
import { AuthModule } from './auth/auth.module';
import { RabbitMQModule } from './messaging/rabbitmq/rabbitmq.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: '../../.env',
    }),
    MailModule,
    AuthModule,
    RabbitMQModule,
  ],
  controllers: [
    MockController,
  ],
})
export class AppModule {}