import { Injectable, InternalServerErrorException } from '@nestjs/common';
import { SearchGeoDto } from './dto/search-geo.dto';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';

@Injectable()
export class GeoSearchService {
  constructor(private readonly qdrantService: QdrantService) {}

  //  Tìm cửa hàng gần người dùng
  async searchLocations(searchGeoDto: SearchGeoDto) {
    try {
      console.log('🚀 [GeoSearchService] Processing search data:', searchGeoDto);
      
      const { lat, lng, radius = 5000 } = searchGeoDto;

      // Gọi xuống tầng hạ tầng Qdrant để lọc tọa độ cửa hàng
      const stores = await this.qdrantService.findStoresByRadius(lat, lng, radius);

      return {
        success: true,
        code: 'UC-P3-GEO-01-SUCCESS',
        message: `Found ${stores.length} stores within ${radius}m`,
        data: stores,
      };
    } catch (error) {
      console.error(' Error in GeoSearchService:', error);
      throw new InternalServerErrorException('Có lỗi xảy ra khi truy vấn dữ liệu vị trí.');
    }
  }
  // UC-P3-GEO-02: Tìm trung tâm bảo hành gần người dùng
async searchServiceCenters(searchGeoDto: SearchGeoDto) {
  try {
    console.log('🚀 [GeoSearchService] Processing service center search:', searchGeoDto);
    
    // Trạm bảo hành thường ít hơn cửa hàng, mặc định cho bán kính quét xa hơn (10km = 10000m)
    const { lat, lng, radius = 10000 } = searchGeoDto;

    // Gọi xuống Qdrant lọc theo type là 'service_center'
    const centers = await this.qdrantService.findLocationsByRadius(lat, lng, radius, 'service_center');

    return {
      success: true,
      message: `Found ${centers.length} service centers within ${radius}m`,
      data: centers,
    };
  } catch (error) {
    console.error('[GeoSearchService] Error in searchServiceCenters:', error);
    throw new InternalServerErrorException('Có lỗi xảy ra khi tìm trung tâm bảo hành.');
  }
}
}