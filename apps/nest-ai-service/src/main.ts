import { NestFactory } from '@nestjs/core';
import { MicroserviceOptions, Transport } from '@nestjs/microservices';

import { AppModule } from './app.module';
import { RABBITMQ } from './messaging/rabbitmq/rabbitmq.constants';

async function bootstrap() {
  const app = await NestFactory.createMicroservice<MicroserviceOptions>(
    AppModule,
    {
      transport: Transport.RMQ,
      options: {
        urls: [RABBITMQ.URL],

        queue: RABBITMQ.QUEUE,

        queueOptions: {
          durable: true,

          arguments: {
            'x-dead-letter-exchange': RABBITMQ.DLX,
            'x-dead-letter-routing-key': RABBITMQ.DLQ_ROUTING_KEY,
          },
        },

        noAck: false,
      },
    },
  );

  await app.listen();

  console.log('Nest AI consumer started');
}

bootstrap();