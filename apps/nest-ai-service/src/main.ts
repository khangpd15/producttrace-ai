import { NestFactory } from '@nestjs/core';
import { MicroserviceOptions, Transport } from '@nestjs/microservices';

import { AppModule } from './app.module';
import { KAFKA } from './kafka/kafka.constants';

async function bootstrap() {
  const app = await NestFactory.createMicroservice<MicroserviceOptions>(AppModule, {
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

  await app.listen();
  console.log('Nest AI Kafka consumer started');
}

bootstrap();