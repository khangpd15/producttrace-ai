import { Controller, Get, Query, UsePipes, ValidationPipe } from '@nestjs/common';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';
import { SearchGeoDto } from './dto/search-geo.dto';

@Controller('geo-search')
export class GeoSearchController {
  constructor(private readonly qdrantService: QdrantService) {}

  @Get('nearest')
  @UsePipes(new ValidationPipe({ transform: true })) 
  async getNearestStore(@Query() query: SearchGeoDto) {
    const { lat, lng, radius = 5000 } = query;

    console.log(`[🚀 UC-P3-GEO-01] Request nearest store: Lat ${lat}, Lng ${lng}, Radius ${radius}m`);

    const stores = await this.qdrantService.findStoresByRadius(lat, lng, radius);

    return {
      success: true,
      message: `Found ${stores.length} stores within ${radius}m`,
      data: stores,
    };
  }
}