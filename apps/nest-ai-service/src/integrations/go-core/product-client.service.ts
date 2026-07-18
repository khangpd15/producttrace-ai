import { Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';

// Full product detail returned by GET /api/products/:id
export interface ProductDetail {
  id: string;
  category_id?: string;
  name: string;
  slug?: string;
  description?: string;
  thumbnail_url?: string;
  tags?: string[];
  metadata?: Record<string, unknown>;
  status?: string;
  created_by?: string;
  created_at?: string;
  updated_at?: string;
  variants?: ProductVariantDetail[];
}

export interface ProductVariantDetail {
  id: string;
  sku: string;
  name: string;
  barcode?: string;
  price?: number;
  currency?: string;
  images?: string[];
  status?: string;
  attributes?: ProductAttributeDetail[];
  created_at?: string;
  updated_at?: string;
}

export interface ProductAttributeDetail {
  id: string;
  attribute_id: string;
  label: string;
  value_text?: string;
  value_number?: number;
  value_boolean?: boolean;
}

// Minimal product item from list API (GET /api/products)
interface ProductListItem {
  id: string;
  name: string;
  category_name: string;
  variants_count: number;
  batches_count: number;
  status?: string;
  created_at: string;
  thumbnail_url?: string;
}

// Go API response wrappers
interface PaginationMeta {
  current_page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
}

interface GoProductListData {
  items: ProductListItem[];
  meta: PaginationMeta;
}

interface ProductListResponse {
  data: ProductListItem[];
  total_pages: number;
}

interface ApiListResponse {
  success: boolean;
  message: string;
  data: GoProductListData;
}

interface ApiDetailResponse {
  success: boolean;
  message: string;
  data: ProductDetail;
}

@Injectable()
export class ProductClientService {
  private readonly logger = new Logger(ProductClientService.name);

  constructor(
    private readonly configService: ConfigService,
  ) { }

  /**
   * Fetches a page of product IDs and basic info from the list API.
   * The list API returns minimal data (no variants, tags, metadata, etc.).
   * For full product data, use getProductById().
   */
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

    const body = (await response.json()) as ApiListResponse;

    return {
      data: Array.isArray(body?.data?.items) ? body.data.items : [],
      total_pages: body?.data?.meta?.total_pages ?? 0,
    };
  }

  /**
   * Fetches full product detail including variants, attributes, tags, metadata.
   * This is the API that returns the same level of detail as the product.created event.
   */
  async getProductById(id: string): Promise<ProductDetail | null> {
    const baseUrl =
      this.configService.get<string>('GO_CORE_URL');

    if (!baseUrl) {
      throw new Error('GO_CORE_URL is not configured');
    }

    const response = await fetch(
      `${baseUrl}/api/products/${id}`,
    );

    if (!response.ok) {
      if (response.status === 404) {
        this.logger.warn(`Product ${id} not found`);
        return null;
      }
      throw new Error(
        `Failed to fetch product ${id}: ${response.status}`,
      );
    }

    const body = (await response.json()) as ApiDetailResponse;

    if (!body?.data) {
      this.logger.warn(`Product ${id}: empty response data`);
      return null;
    }

    return body.data;
  }
}