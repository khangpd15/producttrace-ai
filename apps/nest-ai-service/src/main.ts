import { NestFactory } from '@nestjs/core';
import { ValidationPipe } from '@nestjs/common';
import { MicroserviceOptions, Transport } from '@nestjs/microservices';
import * as dotenv from 'dotenv';
import * as path from 'path';
import * as fs from 'fs';

const envPath = path.join(process.cwd(), '.env');

console.log('ENV PATH =', envPath);
console.log('ENV EXISTS =', fs.existsSync(envPath));

dotenv.config({ path: envPath });

console.log('KAFKA_BROKER AFTER DOTENV =', process.env.KAFKA_BROKER);

import { AppModule } from './app.module';
import { KAFKA } from './kafka/kafka.constants';
import { RABBITMQ } from './messaging/rabbitmq/rabbitmq.constants';

async function bootstrap() {
  try {
    console.log('========== BOOTSTRAP START ==========');

    console.log('Creating Nest application...');
    const app = await NestFactory.create(AppModule);
    console.log('Nest application created.');

    app.enableCors();
    app.useGlobalPipes(new ValidationPipe());
    console.log('Connecting RabbitMQ microservice...');
    console.log('RabbitMQ microservice connected.');

    console.log('Connecting Kafka microservice...');
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
    console.log('Kafka microservice connected.');

    console.log('========== BEFORE startAllMicroservices ==========');
    await app.startAllMicroservices();
    console.log('========== AFTER startAllMicroservices ==========');

    const port = Number(process.env.PORT) || 3000;

    console.log(`========== BEFORE app.listen(${port}) ==========`);

    await app.listen(port, '0.0.0.0');

    console.log('========== AFTER app.listen ==========');
    console.log(`Nest AI Service is running on HTTP port ${port}`);
    console.log('RabbitMQ consumer started');
    console.log('Kafka consumer started');
  } catch (error) {
    console.error('========== BOOTSTRAP ERROR ==========');
    console.error(error);

    if (error instanceof Error) {
      console.error('Message:', error.message);
      console.error('Stack:', error.stack);
    }
  }
}

bootstrap();