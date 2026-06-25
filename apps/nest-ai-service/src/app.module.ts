import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';

import { ProductCreatedConsumer } from './messaging/consumers/product-created.consumer';
import { UserRegisteredConsumer } from './messaging/consumers/user-registered.consumer';
import { MailModule } from './modules/mail/mail.module';

@Module({
  imports: [
    ConfigModule.forRoot({ isGlobal: true }),
    MailModule,
  ],
  controllers: [ProductCreatedConsumer, UserRegisteredConsumer],
})
export class AppModule {}