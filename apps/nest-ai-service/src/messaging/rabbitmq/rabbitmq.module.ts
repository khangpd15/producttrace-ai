import { Module } from '@nestjs/common';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { ConfigModule } from '@nestjs/config';

import { MailModule } from '../../modules/mail/mail.module';
import { RabbitMQService } from './rabbitmq.service';
import { NotificationConsumer } from '../consumers/notification.consumer';
import { RabbitMQProducerService } from './rabbitmq-producer.service';

import { RABBITMQ } from './rabbitmq.constants';

@Module({
  imports: [
    ConfigModule,
    MailModule,
    
     ClientsModule.register([
      {
        name: 'RABBITMQ_PRODUCER',
        transport: Transport.RMQ,
        options: {
          urls: [RABBITMQ.URL],
          exchange: RABBITMQ.EXCHANGE,
          exchangeType: RABBITMQ.EXCHANGE_TYPE,

          queue: RABBITMQ.QUEUES.EMBEDDING,

          queueOptions: {
            durable: true,
            arguments: {
              'x-dead-letter-exchange':
                RABBITMQ.DLX.EMBEDDING,

              'x-dead-letter-routing-key':
                RABBITMQ.DLQ_ROUTING_KEYS.EMBEDDING,
            },
          },
        },
      },
    ]),
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
