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

// Go API thực tế trả về:
// { success, message, data: { items: [...], meta: { current_page, page_size, total_items, total_pages } } }
interface PaginationMeta {
  current_page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
}

interface GoProductListData {
  items: Product[];   // Go dùng "items" không phải "data"
  meta: PaginationMeta;
}

// ProductListResponse là shape mà ReindexService kỳ vọng
interface ProductListResponse {
  data: Product[];
  total_pages: number;
}

interface ApiResponse {
  success: boolean;
  message: string;
  data: GoProductListData;
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

    // Map Go API response shape → shape ReindexService expects
    return {
      data: Array.isArray(body?.data?.items) ? body.data.items : [],
      total_pages: body?.data?.meta?.total_pages ?? 0,
    };
  }
}