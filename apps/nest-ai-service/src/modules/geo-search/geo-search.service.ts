import { Injectable } from '@nestjs/common';
import { SearchGeoDto } from './dto/search-geo.dto';

@Injectable()
export class GeoSearchService {
  async searchLocations(searchGeoDto: SearchGeoDto) {
    // Tạm thời log ra để test xem đã nhận được dữ liệu chưa
    console.log('Search data received:', searchGeoDto);
    
    return {
      message: 'Service is working!',
      data: searchGeoDto,
    };
  }
}