import { Module } from '@nestjs/common';
import { GeoSearchController } from './geo-search.controller';
import { GeoSearchService } from './geo-search.service';
import { QdrantModule } from '../../integrations/qdrant/qdrant.module';

@Module({
  imports: [QdrantModule],
  controllers: [GeoSearchController],
  providers: [GeoSearchService],
})
export class GeoSearchModule {}