import { Controller, Get, Query, UsePipes, ValidationPipe } from '@nestjs/common';
import { GeoSearchService } from './geo-search.service';
import { SearchGeoDto } from './dto/search-geo.dto';

@Controller('geo-search')
export class GeoSearchController {
  constructor(private readonly geoSearchService: GeoSearchService) {}

  @Get('nearest')
  @UsePipes(new ValidationPipe({ transform: true }))
  async getNearestStore(@Query() query: SearchGeoDto) {
    console.log(`[GET] Request nearest store received at Controller: Lat ${query.lat}, Lng ${query.lng}`);
    
    return await this.geoSearchService.searchLocations(query);
  }
}