import { Injectable, InternalServerErrorException } from '@nestjs/common';
import { SearchGeoDto } from './dto/search-geo.dto';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';

@Injectable()
export class GeoSearchService {
  // Inject QdrantService vào để sử dụng kết nối DB
  constructor(private readonly qdrantService: QdrantService) {}

  async searchLocations(searchGeoDto: SearchGeoDto) {
    try {
      console.log('🚀 [GeoSearchService] Processing search data:', searchGeoDto);
      
      const { lat, lng, radius = 5000 } = searchGeoDto;

      // 1. Gọi xuống tầng hạ tầng Qdrant để lọc tọa độ thô từ DB ra
      const stores = await this.qdrantService.findStoresByRadius(lat, lng, radius);

      // 2. Trả kết quả chuẩn hóa về cho Controller
      return {
        success: true,
        message: `Found ${stores.length} stores within ${radius}m`,
        data: stores,
      };
    } catch (error) {
      console.error('❌ Error in GeoSearchService:', error);
      throw new InternalServerErrorException('There was an error while querying the location data.');
    }
  }
}