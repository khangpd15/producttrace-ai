import { Module } from '@nestjs/common';

import { SearchController } from './search.controller';
import { SearchService } from './search.service';
import { RankingService } from './services/ranking.service';

import { QdrantModule } from '../../integrations/qdrant/qdrant.module';
import { EmbeddingModule } from '../embedding/embedding.module';

@Module({
  imports: [
    EmbeddingModule,
    QdrantModule,
  ],
  controllers: [
    SearchController,
  ],
  providers: [
    SearchService,
    RankingService,
  ],
})
export class SearchModule {}