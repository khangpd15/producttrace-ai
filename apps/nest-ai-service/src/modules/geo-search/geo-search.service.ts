import { Injectable, InternalServerErrorException } from '@nestjs/common';
import { SearchStoreDto } from './dto/search-store.dto';
import { SearchServiceCenterDto } from './dto/search-service-center.dto';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';

@Injectable()
export class GeoSearchService {
  constructor(private readonly qdrantService: QdrantService) { }

  // HÀM: Tính khoảng cách giữa 2 tọa độ (Trả về số mét)
   private calculateDistance(lat1: number, lon1: number, lat2: number, lon2: number): number {
    const R = 6371e3; // Bán kính Trái Đất tính bằng mét
    const phi1 = (lat1 * Math.PI) / 180;
    const phi2 = (lat2 * Math.PI) / 180;
    const deltaPhi = ((lat2 - lat1) * Math.PI) / 180;
    const deltaLambda = ((lon2 - lon1) * Math.PI) / 180;

    const a =
      Math.sin(deltaPhi / 2) * Math.sin(deltaPhi / 2) +
      Math.cos(phi1) * Math.cos(phi2) * Math.sin(deltaLambda / 2) * Math.sin(deltaLambda / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));

    return Math.round(R * c);
  }

  // HÀM: Xử lý map dữ liệu, tính khoảng cách, gán note và sắp xếp từ gần đến xa
  private processAndSortLocations(rawLocations: any, userLat: number, userLng: number) {
    const points = Array.isArray(rawLocations) 
      ? rawLocations 
      : (rawLocations?.points || []);

    console.log(`[DEBUG] Tổng số bản ghi thô nhận từ Qdrant: ${points.length}`);

    const processed = points
      .map((point: any) => {
        const payload = point?.payload;
        if (!payload || !payload.location) return null;

        const storeLat = payload.location.lat;
        const storeLng = payload.location.lon;

        const distanceInMeters = this.calculateDistance(userLat, userLng, storeLat, storeLng);

        const distanceNote =
          distanceInMeters >= 1000
            ? `Cách vị trí của bạn ${(distanceInMeters / 1000).toFixed(2)} km`
            : `Cách vị trí của bạn ${distanceInMeters} mét`;

        return {
          id: point.id,
          name: payload.name || 'Địa điểm không rõ tên',
          type: payload.type || 'unknown',
          address: payload.address || 'Không có địa chỉ',
          products: payload.products || undefined,
          distanceMeters: distanceInMeters, 
          note: distanceNote,
        };
      })
      .filter((item: any) => item !== null);

    //  sắp xếp từ gần nhất đến xa nhất
    return [...processed].sort((a: any, b: any) => a.distanceMeters - b.distanceMeters);
  }

  // 1. Tìm tất cả cửa hàng gần người dùng nhất
  async searchLocations(searchStoreDto: SearchStoreDto) {
    try {
      console.log('[GeoSearchService] Processing search stores:', searchStoreDto);
      const { lat, lng, radius = 20000 } = searchStoreDto;

      const rawStores = await this.qdrantService.findStoresByRadius(lat, lng, radius);

      // Áp dụng tính khoảng cách và xếp hạng từ gần đến xa
      const sortedStores = this.processAndSortLocations(rawStores, lat, lng);

      return {
        success: true,
        message: sortedStores.length > 0
          ? `Đã tìm thấy và sắp xếp ${sortedStores.length} cửa hàng trong bán kính ${radius / 1000}km`
          : `Không tìm thấy cửa hàng nào trong bán kính ${radius / 1000}km quanh vị trí của bạn.`,
        data: sortedStores,
      };
    } catch (error) {
      console.error('Error in searchLocations:', error);
      throw new InternalServerErrorException('Có lỗi xảy ra khi truy vấn dữ liệu vị trí.');
    }
  }

  // 2. Tìm trung tâm bảo hành gần người dùng nhất
  async searchServiceCenters(searchServiceCenterDto: SearchServiceCenterDto) {
    try {
      console.log('[GeoSearchService] Processing service center search:', searchServiceCenterDto);
      const { lat, lng, radius = 20000 } = searchServiceCenterDto;

      const rawCenters = await this.qdrantService.findLocationsByRadius(lat, lng, radius, 'service_center');

      // Áp dụng tính khoảng cách và xếp hạng từ gần đến xa
      const sortedCenters = this.processAndSortLocations(rawCenters, lat, lng);

      return {
        success: true,
        message: sortedCenters.length > 0
          ? `Đã tìm thấy và sắp xếp ${sortedCenters.length} trung tâm bảo hành trong bán kính ${radius / 1000}km`
          : `Không tìm thấy trung tâm bảo hành nào trong bán kính ${radius / 1000}km quanh vị trí của bạn.`,
        data: sortedCenters,
      };
    } catch (error) {
      console.error('Error in searchServiceCenters:', error);
      throw new InternalServerErrorException('Có lỗi xảy ra khi tìm trung tâm bảo hành.');
    }
  }

  // 3. Tìm cửa hàng có sản phẩm cụ thể gần người dùng nhất
  async searchProductsNearby(searchStoreDto: SearchStoreDto) {
    try {
      console.log('[GeoSearchService] Processing products nearby search:', searchStoreDto);
      const { lat, lng, radius = 20000, keyword } = searchStoreDto;

      const rawLocations = await this.qdrantService.findProductsByRadius(lat, lng, radius, keyword);

      // Áp dụng tính khoảng cách và xếp hạng từ gần đến xa
      const sortedProductStores = this.processAndSortLocations(rawLocations, lat, lng);

      return {
        success: true,
        message: sortedProductStores.length > 0
          ? `Đã tìm thấy và sắp xếp ${sortedProductStores.length} địa điểm bán [${keyword}] trong bán kính ${radius / 1000}km`
          : `Không tìm thấy cửa hàng nào bán sản phẩm [${keyword}] trong bán kính ${radius / 1000}km lân cận bạn.`,
        data: sortedProductStores,
      };
    } catch (error) {
      console.error('Error in searchProductsNearby:', error);
      throw new InternalServerErrorException('Có lỗi xảy ra khi tìm sản phẩm xung quanh.');
    }
  }
}