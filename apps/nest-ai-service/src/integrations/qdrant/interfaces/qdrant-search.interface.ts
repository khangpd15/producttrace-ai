import { SearchFilter } from '../../../modules/search/interfaces/search-filter.interface';

export interface QdrantSearchRequest {
    vector: number[];
    limit: number;
    filter?: SearchFilter;
}

export interface QdrantSearchResult {
  id: string;
  score: number;
  payload?: Record<string, unknown>;
}