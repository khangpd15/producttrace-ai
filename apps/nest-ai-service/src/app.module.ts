import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';

import { ProductCreatedConsumer } from './messaging/consumers/product-created.consumer';
import { UserRegisteredConsumer } from './messaging/consumers/user-registered.consumer';
import { PasswordResetConsumer } from './messaging/consumers/password-reset.consumer';
import { MailModule } from './modules/mail/mail.module';
import { MockController } from './mock.controller';
import { AuthModule } from './auth/auth.module';
import { SearchModule } from './modules/search/search.module';

@Module({
  imports: [
    ConfigModule.forRoot({
      isGlobal: true,
      envFilePath: '../../.env',
    }),
    MailModule,
    AuthModule,
    SearchModule,
  ],
  controllers: [ProductCreatedConsumer, UserRegisteredConsumer, PasswordResetConsumer, MockController],
})
export class AppModule {}