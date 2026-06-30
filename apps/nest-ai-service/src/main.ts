import { NestFactory } from '@nestjs/core';
import { ValidationPipe } from '@nestjs/common';
import { MicroserviceOptions, Transport } from '@nestjs/microservices';
import * as dotenv from 'dotenv';
import * as path from 'path';

// Load .env before any other imports that might rely on process.env
dotenv.config({ path: path.join(__dirname, '../../../.env') });

import { AppModule } from './app.module';
import { KAFKA } from './kafka/kafka.constants';
import { RABBITMQ } from './messaging/rabbitmq/rabbitmq.constants';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  app.enableCors();
  app.useGlobalPipes(new ValidationPipe());

  // RabbitMQ microservice
  app.connectMicroservice<MicroserviceOptions>({
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
  });

  // Kafka microservice
  app.connectMicroservice<MicroserviceOptions>({
    transport: Transport.KAFKA,
    options: {
      client: {
        clientId: 'nest-ai-service',
        brokers: KAFKA.BROKERS,
      },
      consumer: {
        groupId: KAFKA.GROUP_ID,
      },
    },
  });

  await app.startAllMicroservices();

  const port = process.env.PORT || 3000;
  await app.listen(port);

  console.log(`Nest AI Service is running on HTTP port ${port}`);
  console.log('RabbitMQ consumer started');
  console.log('Kafka consumer started');
}

bootstrap();