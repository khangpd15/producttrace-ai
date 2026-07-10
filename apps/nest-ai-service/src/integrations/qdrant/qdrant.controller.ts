import { Controller, Post } from '@nestjs/common';
import { QdrantService } from './qdrant.service';

@Controller('qdrant')
export class QdrantController {
  constructor(private readonly qdrantService: QdrantService) {}

  // =========================================================================
  // SEED DỮ LIỆU CỬA HÀNG (GỒM THỰC TẾ + GIẢ LẬP )
  // =========================================================================
  @Post('seed-stores')
  async seedStores() {
    const client = (this.qdrantService as any).client;
    const collectionName = (this.qdrantService as any).collectionName;

    // 1. Mảng dữ liệu thực tế 
    const realStores = [
      {
        id: 101,
        name: 'Cửa hàng FPT Shop Cần Thơ',
        type: 'store',
        location: { lat: 10.0226, lon: 105.7314 },
        address: '95-97-99 Hùng Vương, Thới Bình, Ninh Kiều, Cần Thơ',
      },
      {
        id: 102,
        name: 'Đại lý Ủy quyền ProductTrace Vĩnh Long',
        type: 'store',
        location: { lat: 10.1350, lon: 105.9620 },
        address: 'Ba Càng, Song Phú, Tam Bình, Vĩnh Long',
      },
      {
        id: 103,
        name: 'Thế Giới Di Động Phường 1 Vĩnh Long',
        type: 'store',
        location: { lat: 10.2524, lon: 105.9612 },
        address: 'Trần Hưng Đạo, Phường 1, TP. Vĩnh Long',
      },
      {
        id: 104,
        name: 'Showroom ProductTrace Landmark 81',
        type: 'store',
        location: { lat: 10.7948, lon: 106.7218 },
        address: 'Vinhomes Central Park, Bình Thạnh, TP. Hồ Chí Minh',
      },
    ];

    await client.upsert(collectionName, {
      wait: true,
      points: realStores.map(store => ({
        id: store.id,
        payload: {
          name: store.name,
          type: store.type,
          location: store.location,
          address: store.address,
        },
        vector: {},
      })),
    });

    // 2. Thuật toán giả lập tự động tăng tọa độ để tạo ra 10 cửa hàng giả lập(không trùng với thực tế)
    const mockStoresOld: any[] = [];
    const baseLat = 10.0;
    const baseLng = 105.0;

    for (let i = 1; i <= 10; i++) {
      mockStoresOld.push({
        id: 1000 + i,
        payload: {
          name: `Cửa hàng Giả Lập Mẫu Số ${i}`,
          type: 'store',
          location: {
            lat: baseLat + i * 0.01,
            lon: baseLng + i * 0.01,
          },
          address: `Địa chỉ giả lập tự động cấp số ${i}, Khu vực ĐBSCL`,
        },
        vector: {},
      });
    }

    await client.upsert(collectionName, {
      wait: true,
      points: mockStoresOld,
    });

    return {
      success: true,
      message: `Đã nạp thành công ${realStores.length} cửa hàng thực tế và ${mockStoresOld.length} cửa hàng giả lập cũ vào Qdrant!`,
    };
  }

  // =========================================================================
  //  SEED DỮ LIỆU TRUNG TÂM BẢO HÀNH THỰC TẾ
  // =========================================================================
  @Post('seed-service-centers')
  async seedServiceCenters() {
    const client = (this.qdrantService as any).client;
    const collectionName = (this.qdrantService as any).collectionName;

    const mockCenters = [
      {
        id: 201,
        name: 'Trung tâm Bảo hành Ủy quyền ProductTrace - Vĩnh Long',
        type: 'service_center',
        location: { lat: 10.2524, lon: 105.9612 },
        address: 'Phường 1, TP. Vĩnh Long, Tỉnh Vĩnh Long',
      },
      {
        id: 202,
        name: 'Trạm Bảo hành Khu vực Trung tâm Cần Thơ',
        type: 'service_center',
        location: { lat: 10.0226, lon: 105.7314 },
        address: 'Quận Ninh Kiều, Thành phố Cần Thơ',
      }
    ];

    await client.upsert(collectionName, {
      wait: true,
      points: mockCenters.map(c => ({
        id: c.id,
        payload: {
          name: c.name,
          type: c.type,
          location: c.location,
          address: c.address
        },
        vector: {},
      })),
    });

    return {
      success: true,
      message: ` Đã nạp thành công ${mockCenters.length} trung tâm bảo hành thực tế vào Qdrant!`,
    };
  }
}