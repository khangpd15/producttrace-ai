import { Injectable, OnModuleInit } from '@nestjs/common';
import { QdrantClient } from '@qdrant/js-client-rest';

@Injectable()
export class QdrantService implements OnModuleInit {
  // Thêm hàm này vào bên trong class QdrantService của bạn
  async findStoresByRadius(lat: number, lng: number, radiusInMeters: number = 5000) {
    try {
      const result = await this.client.scroll('stores_collection', {
        filter: {
          must: [
            {
              geo_radius: {
                key: 'location',
                range: {
                  radius: radiusInMeters,  
                  center: { lat: lat, lon: lng }, // Tọa độ vị trí của người dùng
                },
              },
            },
          ],
        },
        limit: 10, 
        with_payload: true,
      });

      // Trả ra danh sách payload chứa thông tin cửa hàng
      return result.points.map(point => ({
        id: point.id,
        ...(point.payload as any)
      }));
    } catch (error) {
      console.error('Error from Qdrant:', error);
      throw error;
    }
  }
  private client: QdrantClient;
  private readonly collectionName = 'products_collection';
  private readonly storeCollection = 'stores_collection';

  constructor() {
    this.client = new QdrantClient({
      url: process.env.QDRANT_URL || 'http://localhost:6333',
    });
  }

  async onModuleInit() {
    const collections = await this.client.getCollections();
    const exists = collections.collections.find(c => c.name === this.collectionName);

    if (!exists) {
      await this.client.createCollection(this.collectionName, {
        vectors: { size: 768, distance: 'Cosine' }, // size 768 tùy thuộc vào model embedding 
      });
      console.log('Collection created:', this.collectionName);
    }
  }
  async upsertProduct(productData: { id: number | string, vector: number[], metadata: any, name?: string }) {

    const numericId = this.generateNumericId(productData.id);

    await this.client.upsert('products_collection', {
      wait: true,
      points: [
        {
          id: numericId,
          vector: productData.vector,
          payload: {
            name: productData.name,
            originalId: productData.id
          },
        },
      ],
    });
  }
  private generateNumericId(input: string | number): number {
    const str = String(input);
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      hash = ((hash << 5) - hash) + str.charCodeAt(i);
      hash |= 0;
    }
    return Math.abs(hash); 
  }
  async upsertStoreToQdrant(store: { id: number; name: string; lat: number; lng: number; address?: string }) {
  await this.client.upsert(this.storeCollection, {
    wait: true,
    points: [
      {
        id: store.id, 
        vector: {},   
        payload: {
          name: store.name,
          address: store.address || '',
          location: { 
            lat: Number(store.lat), 
            lon: Number(store.lng) 
          }
        },
      },
    ],
  });
}
}