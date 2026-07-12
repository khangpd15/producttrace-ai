import {
  Injectable,
  Logger,
} from '@nestjs/common';

import { HybridSearchDto } from './dto/hybrid-search.dto';
import { RankingService } from './services/ranking.service';

import { EmbeddingService } from '../embedding/embedding.service';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';

import { SearchResult } from './interfaces/search-result.interface';

@Injectable()
export class SearchService {
  private readonly logger = new Logger(SearchService.name);

  constructor(
    private readonly embeddingService: EmbeddingService,
    private readonly qdrantService: QdrantService,
    private readonly rankingService: RankingService,
  ) { }

  async hybridSearch(
    dto: HybridSearchDto,
  ): Promise<SearchResult[]> {
    try {
      this.logger.log(
        `[SEARCH] Query="${dto.query}"`,
      );

      // 1. Generate embedding
      const vector =
        await this.embeddingService.generateEmbedding(
          dto.query,
        );

      this.logger.log(
        `[SEARCH] Embedding generated (dimension=${vector.length})`,
      );

      // 2. Build metadata filter
      const filter = this.buildFilter(dto);

      this.logger.debug(
        `[SEARCH] Metadata filter: ${JSON.stringify(filter)}`,
      );

      // 3. Search Qdrant
      const results =
        await this.qdrantService.vectorSearch(
          vector,
          filter,
          dto.limit ?? 10,
        );

      this.logger.log(
        `[SEARCH] Qdrant returned ${results.length} results`,
      );

      // 4. Ranking
      const rankedResults =
        this.rankingService.rank(results);

      this.logger.log(
        `[SEARCH] Ranking completed`,
      );

      return rankedResults;
    } catch (error) {
      this.logger.error(
        `[SEARCH] Hybrid search failed`,
        error instanceof Error
          ? error.stack
          : JSON.stringify(error),
      );

      throw error;
    }
  }
  private buildFilter(dto: HybridSearchDto) {
    const filter: Record<string, unknown> = {};

    if (dto.category) {
      filter.category = dto.category;
    }

    if (dto.manufacturer) {
      filter.manufacturer = dto.manufacturer;
    }

    if (dto.province) {
      filter.province = dto.province;
    }

    return Object.keys(filter).length > 0
      ? filter
      : undefined;
  }
}