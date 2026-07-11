import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';

import { MailModule } from '../../modules/mail/mail.module';
import { RabbitMQService } from './rabbitmq.service';
import { NotificationConsumer } from '../consumers/notification.consumer';

@Module({
  imports: [ConfigModule, MailModule],
  providers: [
    RabbitMQService,
    NotificationConsumer,
  ],
  exports: [RabbitMQService],
})
export class RabbitMQModule {}
