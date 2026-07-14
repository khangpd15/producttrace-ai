import { Injectable, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import * as fs from 'fs';
import { SearchResult } from '../../modules/search/interfaces/search-result.interface';

@Injectable()
export class QdrantService {
  private readonly logger = new Logger(QdrantService.name);

  private readonly baseUrl: string;
  private readonly apiKey: string | undefined;

  // ✅ COLLECTION NAME (fixed)
  private readonly collectionName = 'product_embeddings';

  // ✅ FIXED VECTOR SIZE (quan trọng cho production)
  private readonly vectorSize = 768; // đổi theo model của bạn

  private collectionReady = false;
  private creatingCollection = false;

  constructor(private readonly configService: ConfigService) {
    const isDocker = fs.existsSync('/.dockerenv');
    const defaultBaseUrl = isDocker ? 'http://qdrant:6333' : 'http://localhost:6333';
    let envBaseUrl = this.configService.get<string>('QDRANT_URL');
    if (envBaseUrl && envBaseUrl.trim() === '') {
      envBaseUrl = undefined;
    }
    const resolvedBaseUrl = !isDocker && envBaseUrl?.includes('qdrant')
      ? defaultBaseUrl
      : envBaseUrl ?? defaultBaseUrl;
    this.baseUrl = resolvedBaseUrl;
    this.apiKey = this.configService.get('QDRANT_API_KEY');
  }

  // =========================
  // UPSERT VECTOR
  // =========================
  async upsertVector(
    id: string,
    vector: number[],
    payload: Record<string, unknown>,
  ): Promise<void> {
    // validate vector
    if (!vector || vector.length !== this.vectorSize) {
      throw new Error(
        `Invalid vector size: expected ${this.vectorSize}, got ${vector?.length}`,
      );
    }

    // ensure collection exists
    await this.ensureCollection();

    const response = await fetch(
      `${this.baseUrl}/collections/${this.collectionName}/points?wait=true`,
      {
        method: 'PUT',
        headers: this.headers,
        body: JSON.stringify({
          points: [
            {
              id,
              vector,
              payload,
            },
          ],
        }),
      },
    );

    if (!response.ok) {
      const body = await response.text();
      this.logger.error(`Qdrant upsert failed: ${response.status} ${body}`);
      throw new Error(`Qdrant upsert failed: ${response.status}`);
    }

    this.logger.log(
      `Upsert success: id=${id}, collection=${this.collectionName}`,
    );
  }

  // =========================
  // VECTOR SEARCH
  // =========================
  async vectorSearch(
    vector: number[],
    filter?: Record<string, unknown>,
    limit = 10,
  ): Promise<SearchResult[]> {
    // validate vector
    if (!vector || vector.length !== this.vectorSize) {
      throw new Error(
        `Invalid vector size: expected ${this.vectorSize}, got ${vector?.length}`,
      );
    }

    // ensure collection exists
    await this.ensureCollection();

    const response = await fetch(
      `${this.baseUrl}/collections/${this.collectionName}/points/search`,
      {
        method: 'POST',
        headers: this.headers,
        body: JSON.stringify({
          vector,
          filter,
          limit,
          with_payload: true,
          with_vector: false,
        }),
      },
    );

    if (!response.ok) {
      const body = await response.text();

      this.logger.error(
        `Qdrant search failed: ${response.status} ${body}`,
      );

      throw new Error(
        `Qdrant search failed: ${response.status}`,
      );
    }

    const body = await response.json();

    return body.result as SearchResult [];
  }
  // =========================
  // ENSURE COLLECTION
  // =========================
  private async ensureCollection(): Promise<void> {
    if (this.collectionReady) return;
    if (this.creatingCollection) return;

    this.creatingCollection = true;

    try {
      const exists = await this.collectionExists();

      if (!exists) {
        this.logger.log(
          `Creating collection=${this.collectionName}, vectorSize=${this.vectorSize}`,
        );

        const response = await fetch(
          `${this.baseUrl}/collections/${this.collectionName}`,
          {
            method: 'PUT',
            headers: this.headers,
            body: JSON.stringify({
              vectors: {
                size: this.vectorSize,
                distance: 'Cosine',
              },
            }),
          },
        );

        if (!response.ok) {
          const body = await response.text();
          throw new Error(
            `Qdrant collection creation failed: ${response.status} ${body}`,
          );
        }

        this.logger.log(`Collection created: ${this.collectionName}`);
      }

      this.collectionReady = true;
    } finally {
      this.creatingCollection = false;
    }
  }

  // =========================
  // CHECK COLLECTION EXISTS
  // =========================
  private async collectionExists(): Promise<boolean> {
    const response = await fetch(
      `${this.baseUrl}/collections/${this.collectionName}`,
      {
        method: 'GET',
        headers: this.headers,
      },
    );

    if (response.ok) return true;

    if (response.status === 404) return false;

    const body = await response.text();
    throw new Error(
      `Qdrant check failed: ${response.status} ${body}`,
    );
  }

  // =========================
  // HEADERS
  // =========================
  private get headers() {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (this.apiKey) {
      headers['api-key'] = this.apiKey;
    }

    return headers;
  }
}