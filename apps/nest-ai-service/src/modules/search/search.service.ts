import { Injectable } from '@nestjs/common';
import { VectorSearchDto } from './dto/vector-search.dto';
import { SearchResult } from './interfaces/search-result.interface';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';

@Injectable()
export class SearchService {
  constructor(
    private readonly qdrantService: QdrantService,
  ) {}

  private buildSearchRequest(dto: VectorSearchDto) {
    return {
      vector: [], // TODO: Embedding Service sẽ trả vector
      limit: dto.limit ?? 10,
      filter: dto.filter,
    };
  }

  async vectorSearch(
    dto: VectorSearchDto,
  ): Promise<SearchResult[]> {

    console.log('Search query:', dto.query);

    const request = this.buildSearchRequest(dto);

    const result = await this.qdrantService.search(request);

    if (!result.length) {
      return [];
    }

    return result;
  }
}