import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { MailModule } from '../../modules/mail/mail.module';
import { RabbitMQService } from './rabbitmq.service';
import { UserRegisteredConsumer } from '../consumers/user-registered.consumer';
import { UserVerifiedConsumer } from '../consumers/user-verified.consumer';
import { PasswordResetConsumer } from '../consumers/password-reset.consumer';
import { ProductCreatedConsumer } from '../consumers/product-created.consumer';

@Module({
  imports: [ConfigModule, MailModule],
  providers: [
    RabbitMQService,
    UserRegisteredConsumer,
    UserVerifiedConsumer,
    PasswordResetConsumer,
    ProductCreatedConsumer,
  ],
  exports: [RabbitMQService],
})
export class RabbitMQModule {}
