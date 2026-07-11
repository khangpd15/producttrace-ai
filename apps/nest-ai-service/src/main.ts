import { NestFactory } from '@nestjs/core';
import { ValidationPipe } from '@nestjs/common';
import {
  MicroserviceOptions,
  Transport,
} from '@nestjs/microservices';
import * as dotenv from 'dotenv';
import * as path from 'path';
import * as amqp from 'amqplib';

dotenv.config({
  path: path.join(__dirname, '../../../.env'),
});
console.log('ENV URL =', process.env.RABBITMQ_URL);

import { AppModule } from './app.module';
import { RABBITMQ } from './integrations/rabbitmq/rabbitmq.constants';
console.log('CONST URL =', RABBITMQ.URL);

async function ensureEmbeddingTopology() {
  const connection = await amqp.connect(RABBITMQ.URL);
  const channel = await connection.createChannel();

  try {
    await channel.assertExchange(
      RABBITMQ.EXCHANGE,
      RABBITMQ.EXCHANGE_TYPE,
      { durable: true },
    );

    await channel.assertExchange(RABBITMQ.DLX.EMBEDDING, 'topic', { durable: true });

    await channel.assertQueue(RABBITMQ.QUEUES.EMBEDDING, {
      durable: true,
      arguments: {
        'x-dead-letter-exchange': RABBITMQ.DLX.EMBEDDING,
        'x-dead-letter-routing-key': RABBITMQ.DLQ_ROUTING_KEYS.EMBEDDING,
      },
    });

    await channel.assertQueue(RABBITMQ.DLQ_ROUTING_KEYS.EMBEDDING, {
      durable: true,
    });

    await channel.bindQueue(
      RABBITMQ.DLQ_ROUTING_KEYS.EMBEDDING,
      RABBITMQ.DLX.EMBEDDING,
      RABBITMQ.DLQ_ROUTING_KEYS.EMBEDDING,
    );

    for (const routingKey of [
      RABBITMQ.ROUTING_KEYS.PRODUCT_CREATED,
      RABBITMQ.ROUTING_KEYS.TRACE_EVENTS,
    ]) {
      await channel.bindQueue(
        RABBITMQ.QUEUES.EMBEDDING,
        RABBITMQ.EXCHANGE,
        routingKey,
      );
      console.log(
        `Bound ${RABBITMQ.QUEUES.EMBEDDING} ← [${RABBITMQ.EXCHANGE}] ${routingKey}`,
      );
    }
  } finally {
    await channel.close();
    await connection.close();
  }
}

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  app.enableCors();
  app.useGlobalPipes(new ValidationPipe());

  await ensureEmbeddingTopology();

  // RabbitMQ consumer for Embedding (@EventPattern)
  app.connectMicroservice<MicroserviceOptions>({
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