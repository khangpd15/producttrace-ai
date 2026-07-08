export interface ProductPayload {
  name: string;
  brand: string;
}

export interface SearchResult {
  productId: string;
  score: number;
  payload: ProductPayload;
}