import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { ReindexService } from './reindex.service';
import { ReindexConsumer } from './reindex.consumer';
import { ReindexController } from './reindex.controller';
import { ProductClientService } from '../../integrations/go-core/product-client.service';
import { RabbitMQModule } from '../../messaging/rabbitmq/rabbitmq.module';
import { EmbeddingModule } from '../embedding/embedding.module';

@Module({
  imports: [ConfigModule, RabbitMQModule, EmbeddingModule],
  controllers: [
    ReindexConsumer,
    ReindexController,
  ],
  providers: [ReindexService, ProductClientService],
})
export class ReindexModule {}
