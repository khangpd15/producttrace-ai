import { Module } from '@nestjs/common';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { RabbitMQProducerService } from './rabbitmq-producer.service';
import { RABBITMQ } from './rabbitmq.constants';

@Module({
  imports: [
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
    RabbitMQProducerService,
  ],

  exports: [
    RabbitMQProducerService,
  ],
})
export class EmbeddingRabbitMQModule { }