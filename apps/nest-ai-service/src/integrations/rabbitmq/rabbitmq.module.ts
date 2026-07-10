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
          queue: RABBITMQ.QUEUES.AI_EVENTS,
          queueOptions: {
            durable: true,
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
export class RabbitMQModule {}