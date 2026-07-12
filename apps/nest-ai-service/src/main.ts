import { NestFactory } from '@nestjs/core';
import { ValidationPipe } from '@nestjs/common';
import { MicroserviceOptions, Transport } from '@nestjs/microservices';
import * as dotenv from 'dotenv';
import * as path from 'path';

// Load .env before any other imports that might rely on process.env
dotenv.config({ path: path.join(__dirname, '../../../.env') });

import { AppModule } from './app.module';
import { RABBITMQ } from './messaging/rabbitmq/rabbitmq.constants';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);
  app.enableCors();
  app.useGlobalPipes(new ValidationPipe());
  // Nếu biến môi trường là 'rabbitmq' nhưng đang chạy ngoài Docker (local máy thật),
  // hệ thống tự đổi thành 'localhost' để không bị lỗi kết nối.
  let rabbitmqUrl = process.env.RABBITMQ_URL || (RABBITMQ as any).URL || (RABBITMQ as any).uri || 'amqp://admin:admin123@localhost:5672';
  rabbitmqUrl = String(rabbitmqUrl);

  if (rabbitmqUrl.includes('@rabbitmq:')) {
    console.log('[RabbitMQ] Đang chạy app ở local máy thật, tự động chuyển host sang localhost để tránh lỗi kết nối RabbitMQ');
    rabbitmqUrl = rabbitmqUrl.replace('@rabbitmq:', '@localhost:');
  } else if (rabbitmqUrl.includes('//rabbitmq:')) {
    rabbitmqUrl = rabbitmqUrl.replace('//rabbitmq:', '//localhost:');
  }

  // Connect RabbitMQ microservice
  app.connectMicroservice<MicroserviceOptions>({
    transport: Transport.RMQ,
    options: {
      urls: [rabbitmqUrl],
      queue: process.env.RABBITMQ_QUEUE || RABBITMQ.QUEUE,
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

  await app.startAllMicroservices();
  
  const port = process.env.PORT || 3001;
  await app.listen(port);

  console.log(`Nest AI Service is running on HTTP port ${port}`);
  console.log('Nest AI RabbitMQ consumer started');
}

bootstrap();