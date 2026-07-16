import { Controller, Get, Query, UsePipes, ValidationPipe } from '@nestjs/common';
import { GeoSearchService } from './geo-search.service';
import { SearchStoreDto } from './dto/search-store.dto';
import { SearchServiceCenterDto } from './dto/search-service-center.dto';
// Sửa dòng này:
@Controller('api/geo-search') 
export class GeoSearchController {
  constructor(private readonly geoSearchService: GeoSearchService) {}

  // Các route con: 
  // GET /api/geo-search/nearest
  // GET /api/geo-search/nearest-service-centers
  // GET /api/geo-search/products-nearby

  @Get('nearest')
  @UsePipes(new ValidationPipe({ transform: true }))
  async getNearestStore(@Query() query: SearchStoreDto) {
    console.log(`[GET] Request nearest store received: Lat ${query.lat}, Lng ${query.lng}`);
    return await this.geoSearchService.searchLocations(query);
  }

  @Get('nearest-service-centers')
  @UsePipes(new ValidationPipe({ transform: true }))
  async getNearestServiceCenter(@Query() query: SearchServiceCenterDto) {
    return await this.geoSearchService.searchServiceCenters(query);
  }

  @Get('products-nearby')
  @UsePipes(new ValidationPipe({ transform: true }))
  async getProductsNearby(@Query() query: SearchStoreDto) {
    return await this.geoSearchService.searchProductsNearby(query);
  }
}