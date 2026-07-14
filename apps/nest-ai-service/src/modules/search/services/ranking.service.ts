import { Injectable } from '@nestjs/common';
import { SearchResult } from '../interfaces/search-result.interface';

@Injectable()
export class RankingService {

  rank(results: SearchResult[]): SearchResult[] {

    return results.sort((a, b) => b.score - a.score);

  }

}