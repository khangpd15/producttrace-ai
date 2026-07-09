import { NestFactory } from '@nestjs/core';
import { ValidationPipe } from '@nestjs/common';
import { MicroserviceOptions, Transport } from '@nestjs/microservices';
import * as dotenv from 'dotenv';
import * as path from 'path';

dotenv.config({
  path: path.join(__dirname, '../../../.env'),
});

import { AppModule } from './app.module';

async function bootstrap() {

  const app = await NestFactory.create(AppModule);

  app.enableCors();
  app.useGlobalPipes(new ValidationPipe());


  // Start RabbitMQ Consumer
  // Note: We use the custom RabbitMQService to consume events manually.
  // app.connectMicroservice<MicroserviceOptions>({
  //   transport: Transport.RMQ,
  //   options: {
  //     urls: [
  //       process.env.RABBITMQ_URL ||
  //       'amqp://admin:admin123@localhost:5672/%2F'
  //     ],
  //     queue: 'ai.events',
  //     queueOptions: {
  //       durable: true,
  //       arguments: {
  //         'x-dead-letter-exchange': 'product-trace.dlx',
  //         'x-dead-letter-routing-key': 'ai.events.failed',
  //       },
  //     },
  //     noAck: false,
  //   },
  // });

  // await app.startAllMicroservices();


  const port = process.env.PORT || 3000;

  await app.listen(port);


  console.log(`Nest AI Service running on ${port}`);
  console.log(`RabbitMQ consumer listening on ai.events`);
}

bootstrap();