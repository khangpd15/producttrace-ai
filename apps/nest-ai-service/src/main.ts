import { NestFactory } from '@nestjs/core';
import { ValidationPipe } from '@nestjs/common';
import { MicroserviceOptions, Transport } from '@nestjs/microservices';
import * as dotenv from 'dotenv';
import * as path from 'path';

// Load .env before any other imports that might rely on process.env
dotenv.config({ path: path.join(__dirname, '../../../.env') });

import { AppModule } from './app.module';
async function bootstrap() {
  const app = await NestFactory.create(AppModule);
  app.enableCors();
  app.useGlobalPipes(new ValidationPipe());

  const port = process.env.PORT || 3000;
  await app.listen(port);

  console.log(`Nest AI Service is running on HTTP port ${port}`);
  console.log('Nest AI RabbitMQ consumer started via RabbitMQModule');
}

bootstrap();