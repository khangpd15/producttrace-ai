import { Controller, Get, Query } from '@nestjs/common';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';
import { SearchGeoDto } from './dto/search-geo.dto'; 

@Controller('geo')
export class GeoSearchController {
  constructor(private readonly qdrantService: QdrantService) {}

  @Get('/nearest-store')
  async getNearestStore(@Query() query: SearchGeoDto) { 
    console.log(`[GET] Request finding nearest stores: Lat ${query.lat}, Lng ${query.lng}`);

    const stores = await this.qdrantService.findStoresByRadius(
      query.lat, 
      query.lng, 
      query.radius // DTO tự gán mặc định là 5000 
    );

    return {
      success: true,
      message: `Found ${stores.length} stores within ${query.radius}m`, 
      data: stores,
    };
  }
}