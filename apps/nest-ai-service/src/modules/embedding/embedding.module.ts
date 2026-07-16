import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { EmbeddingConsumer } from './embedding.consumer';
import { EmbeddingService } from './embedding.service';
import { RabbitMQModule } from '../../messaging/rabbitmq/rabbitmq.module';

@Module({
  imports: [ConfigModule, RabbitMQModule],
  controllers: [
    EmbeddingConsumer,
  ],
  providers: [EmbeddingService],
  exports: [EmbeddingService],
})
export class EmbeddingModule {}
