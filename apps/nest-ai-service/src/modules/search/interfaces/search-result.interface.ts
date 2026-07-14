export interface SearchResult {
  id: string;

  score: number;

  payload: Record<string, unknown>;
}