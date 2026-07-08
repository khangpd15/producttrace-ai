import { Injectable } from '@nestjs/common';
import { SearchFilter } from '../../modules/search/interfaces/search-filter.interface';
import { SearchResult } from '../../modules/search/interfaces/search-result.interface';
import { ConfigService } from '@nestjs/config';
import { QdrantSearchRequest } from './interfaces/qdrant-search.interface';


@Injectable()
export class QdrantService {
  constructor(
    private readonly configService: ConfigService,
  ) { }

  private getConfig() {

    return {
      url: this.configService.get<string>('QDRANT_URL'),
      apiKey: this.configService.get<string>('QDRANT_API_KEY'),
      collection: this.configService.get<string>('QDRANT_COLLECTION'),
    };

  }

  private buildQdrantFilter(filter?: SearchFilter) {

    if (!filter) {
      return undefined;
    }

    return {
      must: Object.entries(filter).map(([key, value]) => ({
        key,
        match: {
          value,
        },
      })),
    };

  }
  async search(
    request: QdrantSearchRequest,
  ): Promise<SearchResult[]> {

    const filter = this.buildQdrantFilter(request.filter);

    console.log('====== QDRANT SEARCH ======');
    console.log('Vector:', request.vector);
    console.log('Limit:', request.limit);
    console.log('Filter:', JSON.stringify(filter, null, 2));

    return [
      {
        productId: 'SP001',
        score: 0.96, payload:
        {
          name: 'Laptop Dell XPS 13',
          brand: 'Dell',
        },
      },
    ];
  }
}