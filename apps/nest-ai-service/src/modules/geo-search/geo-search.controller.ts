import { Controller, Get, Query, UsePipes, ValidationPipe } from '@nestjs/common';
import { GeoSearchService } from './geo-search.service';
import { SearchGeoDto } from './dto/search-geo.dto';

@Controller() 
export class GeoSearchController {
  constructor(private readonly geoSearchService: GeoSearchService) {}

  //Tìm cửa hàng gần người dùng nhất
  @Get('geo-search/nearest')
  @UsePipes(new ValidationPipe({ transform: true }))
  async getNearestStore(@Query() query: SearchGeoDto) {
    console.log(`[GET] Request nearest store received at Controller: Lat ${query.lat}, Lng ${query.lng}`);
    return await this.geoSearchService.searchLocations(query);
  }

  //Tìm trung tâm bảo hành gần nhất
  @Get('geo-search/nearest-service-centers')
  @UsePipes(new ValidationPipe({ transform: true }))
  async getNearestServiceCenter(@Query() query: SearchGeoDto) {
    console.log(`[GET] Request nearest service center received at Controller: Lat ${query.lat}, Lng ${query.lng}`);
    return await this.geoSearchService.searchServiceCenters(query);
  }

  // Tìm sản phẩm cùng loại trong bán kính R (mét) 
  @Get('geo-search/products-nearby')
  @UsePipes(new ValidationPipe({ transform: true }))
  async getProductsNearby(@Query() query: SearchGeoDto) {
    console.log(`[GET] Request products nearby received at Controller: Lat ${query.lat}, Lng ${query.lng}, Product ${query.keyword}, Radius ${query.radius}`);
    return await this.geoSearchService.searchProductsNearby(query);
  }
}