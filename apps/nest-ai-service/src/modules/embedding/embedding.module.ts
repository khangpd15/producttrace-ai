import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { EmbeddingConsumer } from './embedding.consumer';
import { EmbeddingService } from './embedding.service';
import { ReindexService } from './reindex.service';
import { ReindexConsumer } from './reindex.consumer';
import { ProductClientService } from '../../integrations/go-core/product-client.service';
import { EmbeddingRabbitMQModule } from '../../integrations/rabbitmq/ai-rabbitmq.module';

@Module({
  imports: [ConfigModule, EmbeddingRabbitMQModule],
  controllers: [EmbeddingConsumer, ReindexConsumer],
  providers: [EmbeddingService, ReindexService, ProductClientService],
})
export class EmbeddingModule {}
