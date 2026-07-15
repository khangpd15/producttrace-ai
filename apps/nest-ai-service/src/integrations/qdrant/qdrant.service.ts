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
  private readonly vectorSize = 1024; // đổi theo model BGE-M3

  private collectionReady = false;
  // Singleton Promise: concurrent callers all await the same creation task
  // instead of returning early (the previous bool-flag race condition).
  private ensureCollectionPromise: Promise<void> | null = null;

  constructor(private readonly configService: ConfigService) {
    const isDocker = fs.existsSync('/.dockerenv');
    const defaultBaseUrl = isDocker ? 'http://qdrant:6333' : 'http://localhost:6333';
    const envBaseUrl = this.configService.get('QDRANT_URL');
    const resolvedBaseUrl = (envBaseUrl && envBaseUrl.trim() !== '')
      ? envBaseUrl
      : defaultBaseUrl;
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
    // Fast path — collection already confirmed ready
    if (this.collectionReady) return;

    // Singleton Promise pattern: if creation is already in progress, all
    // concurrent callers await the same Promise instead of returning early
    // (the previous bool-flag approach caused a race condition where the
    // second caller skipped creation and immediately tried to upsert into a
    // non-existent collection).
    if (!this.ensureCollectionPromise) {
      this.ensureCollectionPromise = this.doEnsureCollection().finally(() => {
        this.ensureCollectionPromise = null;
      });
    }

    return this.ensureCollectionPromise;
  }

  private async doEnsureCollection(): Promise<void> {
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
    } else {
      await this.validateCollectionConfig();
    }

    this.collectionReady = true;
  }

  private async validateCollectionConfig(): Promise<void> {
    const response = await fetch(
      `${this.baseUrl}/collections/${this.collectionName}`,
      {
        method: 'GET',
        headers: this.headers,
      },
    );

    if (!response.ok) {
      const body = await response.text();
      throw new Error(
        `Failed to validate Qdrant collection: ${response.status} ${body}`,
      );
    }

    const info = await response.json();
    // Qdrant v1.x response: result.config.params.vectors.{size, distance}
    const vectorsCfg =
      info?.result?.vectors ??
      info?.result?.config?.params?.vectors;

    const currentSize: number | undefined =
      vectorsCfg?.size ?? vectorsCfg?.vector_size;

    // Normalise to lowercase for safe comparison ('Cosine' vs 'cosine')
    const currentDistance: string = String(
      vectorsCfg?.distance ?? vectorsCfg?.distance_metric ?? '',
    ).toLowerCase();

    if (currentSize !== this.vectorSize || currentDistance !== 'cosine') {
      throw new Error(
        `Qdrant collection ${this.collectionName} config mismatch: ` +
        `expected size=${this.vectorSize}, distance=Cosine but found ` +
        `size=${currentSize}, distance=${currentDistance}. ` +
        `Delete and recreate the collection before upserting vectors.`,
      );
    }

    this.logger.log(
      `[QDRANT] Collection config validated: size=${currentSize}, distance=${currentDistance}`,
    );
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