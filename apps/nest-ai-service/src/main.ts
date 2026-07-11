import { NestFactory } from '@nestjs/core';
import { ValidationPipe } from '@nestjs/common';
import {
  MicroserviceOptions,
  Transport,
} from '@nestjs/microservices';
import * as dotenv from 'dotenv';
import * as path from 'path';

dotenv.config({
  path: path.join(__dirname, '../../../.env'),
});

import { AppModule } from './app.module';
import { RABBITMQ } from './integrations/rabbitmq/rabbitmq.constants';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  app.enableCors();
  app.useGlobalPipes(new ValidationPipe());

  // RabbitMQ consumer for Embedding (@EventPattern)
  app.connectMicroservice<MicroserviceOptions>({
    transport: Transport.RMQ,
    options: {
      urls: [RABBITMQ.URL],
      queue: RABBITMQ.QUEUES.AI_EVENTS,
      queueOptions: {
        durable: true,
      },
      noAck: false,
    },
  });

  await app.startAllMicroservices();

  const port = Number(process.env.PORT) || 3000;

  await app.listen(port);

  console.log(`Nest AI Service running on ${port}`);
  console.log(
    `RabbitMQ Embedding consumer listening on ${RABBITMQ.QUEUES.EMBEDDING}`,
  );
}

bootstrap();