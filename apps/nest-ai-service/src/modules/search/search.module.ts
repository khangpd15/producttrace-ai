import { Module } from '@nestjs/common';
import { SearchController } from './search.controller';
import { SearchService } from './search.service';

import { QdrantModule } from '../../integrations/qdrant/qdrant.module';

@Module({
  imports: [QdrantModule],

  controllers: [SearchController],

  providers: [SearchService],
})
export class SearchModule {}