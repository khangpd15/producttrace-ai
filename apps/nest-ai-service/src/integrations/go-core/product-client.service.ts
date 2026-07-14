import { Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';

interface Product {
  id: string;
  category_id?: string;
  name: string;
  description?: string;
  thumbnail_url?: string;
  status?: string;
}

interface ProductListResponse {
  data: Product[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

interface ApiResponse {
  success: boolean;
  message: string;
  data: ProductListResponse;
}

@Injectable()
export class ProductClientService {
  constructor(
    private readonly configService: ConfigService,
  ) { }

  async getProducts(
    page: number,
    limit: number,
  ): Promise<ProductListResponse> {
    const baseUrl =
      this.configService.get<string>('GO_CORE_URL');

    if (!baseUrl) {
      throw new Error('GO_CORE_URL is not configured');
    }

    const response = await fetch(
      `${baseUrl}/api/products?page=${page}&limit=${limit}`,
    );

    if (!response.ok) {
      throw new Error(
        `Failed to fetch products: ${response.status}`,
      );
    }

    const body = (await response.json()) as ApiResponse;

    return body.data;
  }
}