import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';

import { MailModule } from '../../modules/mail/mail.module';
import { RabbitMQService } from './rabbitmq.service';
import { NotificationConsumer } from '../consumers/notification.consumer';
import { RabbitMQProducerService } from './rabbitmq-producer.service';

@Module({
  imports: [
    ConfigModule,
    MailModule,
    // NOTE: ClientsModule removed — RabbitMQProducerService now uses amqplib
    // directly (see rabbitmq-producer.service.ts) so NestJS ClientProxy is no
    // longer needed and was causing messages to be published to the wrong
    // exchange / wrapped in the NestJS { pattern, data } envelope.
  ],

  providers: [
    RabbitMQService,
    RabbitMQProducerService,
    NotificationConsumer,
  ],
  exports: [
    RabbitMQService,
    RabbitMQProducerService,
  ],
})
export class RabbitMQModule { }
