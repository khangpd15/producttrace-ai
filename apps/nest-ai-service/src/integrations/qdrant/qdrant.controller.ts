import { Controller, Post } from '@nestjs/common';
import { QdrantService } from './qdrant.service';

@Controller('qdrant')
export class QdrantController {
  constructor(private readonly qdrantService: QdrantService) {}

  // =========================================================================
  // SEED DỮ LIỆU CỬA HÀNG
  // =========================================================================
  @Post('seed-stores')
  async seedStores() {
    const client = (this.qdrantService as any).client;
    const collectionName = (this.qdrantService as any).collectionName;

    // Danh sách 5 sản phẩm 
    const productPool = [
      'iPhone 15 Pro',
      'MacBook Air M3',
      'Chuột Logitech G304',
      'Bàn phím Cơ Keychron',
      'Tai nghe Sony WH-1000XM5'
    ];

    // Mảng chứa toàn bộ 10 cửa hàng 
    const mixedStores = [
      {
        id: 101, 
        name: 'Cửa hàng FPT Shop Cần Thơ',
        type: 'store',
        location: { lat: 10.0226, lon: 105.7314 },
        address: '95-97-99 Hùng Vương, Thới Bình, Ninh Kiều, Cần Thơ',
        products: [productPool[0], productPool[1], productPool[4]], // iPhone 15 Pro, MacBook Air M3, Tai nghe Sony
      },
      {
        id: 1001, 
        name: 'Cửa hàng Giả Lập Mẫu Số 1',
        type: 'store',
        location: { lat: 10.95, lon: 105.85 },
        address: 'Địa chỉ giả lập tự động cấp số 1, Khu vực ĐBSCL',
        products: [productPool[2], productPool[3]], // Chuột, Bàn phím
      },
      {
        id: 102, 
        name: 'Đại lý Ủy quyền ProductTrace Vĩnh Long',
        type: 'store',
        location: { lat: 10.1350, lon: 105.9620 },
        address: 'Ba Càng, Song Phú, Tam Bình, Vĩnh Long',
        products: [productPool[0], productPool[2]], // iPhone 15 Pro, Chuột
      },
      {
        id: 1002, 
        name: 'Cửa hàng Giả Lập Mẫu Số 2',
        type: 'store',
        location: { lat: 11.20, lon: 106.10 },
        address: 'Địa chỉ giả lập tự động cấp số 2, Khu vực ĐBSCL',
        products: [productPool[1], productPool[4]], // MacBook Air M3, Tai nghe Sony
      },
      {
        id: 103, 
        name: 'Thế Giới Di Động Phường 1 Vĩnh Long',
        type: 'store',
        location: { lat: 10.2524, lon: 105.9612 },
        address: 'Trần Hưng Đạo, Phường 1, TP. Vĩnh Long',
        products: [productPool[0], productPool[1], productPool[3]], // iPhone 15 Pro, MacBook Air M3, Bàn phím
      },
      {
        id: 1003, 
        name: 'Cửa hàng Giả Lập Mẫu Số 3',
        type: 'store',
        location: { lat: 10.55, lon: 105.45 },
        address: 'Địa chỉ giả lập tự động cấp số 3, Khu vực ĐBSCL',
        products: [productPool[2], productPool[3], productPool[4]], // Chuột, Bàn phím, Tai nghe Sony
      },
      {
        id: 104, 
        name: 'Showroom ProductTrace Landmark 81',
        type: 'store',
        location: { lat: 10.7948, lon: 106.7218 },
        address: 'Vinhomes Central Park, Bình Thạnh, TP. Hồ Chí Minh',
        products: [productPool[1], productPool[3]], // MacBook Air M3, Bàn phím
      },
      {
        id: 1004, 
        name: 'Cửa hàng Giả Lập Mẫu Số 4',
        type: 'store',
        location: { lat: 9.95, lon: 106.35 },
        address: 'Địa chỉ giả lập tự động cấp số 4, Khu vực ĐBSCL',
        products: [productPool[0], productPool[4]], // iPhone 15 Pro, Tai nghe Sony
      },
      {
        id: 1005, 
        name: 'Cửa hàng Giả Lập Mẫu Số 5',
        type: 'store',
        location: { lat: 10.35, lon: 105.75 },
        address: 'Địa chỉ giả lập tự động cấp số 5, Khu vực ĐBSCL',
        products: [productPool[0], productPool[1], productPool[2]], // iPhone 15 Pro, MacBook Air M3, Chuột
      },
      {
        id: 1006,
        name: 'Cửa hàng Giả Lập Mẫu Số 6',
        type: 'store',
        location: { lat: 10.18, lon: 106.25 },
        address: 'Địa chỉ giả lập tự động cấp số 6, Khu vực ĐBSCL',
        products: [productPool[3], productPool[4]], // Bàn phím, Tai nghe Sony
      }
    ];

    // Đẩy nguyên cụm mảng này lên Qdrant
    await client.upsert(collectionName, {
      wait: true,
      points: mixedStores.map(store => ({
        id: store.id,
        payload: {
          name: store.name,
          type: store.type,
          location: store.location,
          address: store.address,
          products: store.products, // 
        },
        vector: {},
      })),
    });

    return {
      success: true,
      message: `Đã nạp thành công 10 cửa hàng (Thực tế + Giả lập) với thứ tự lộn xộn và sẵn sàng để test thuật toán sắp xếp!`,
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
      message: `Đã nạp thành công ${mockCenters.length} trung tâm bảo hành thực tế vào Qdrant!`,
    };
  }
}