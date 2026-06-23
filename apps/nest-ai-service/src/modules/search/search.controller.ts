// src/modules/search/search.controller.ts
import { Controller, Post, Body, HttpCode, HttpStatus } from '@nestjs/common';
import { UnifiedSearchDto } from './/dto/unified-search.dto';

@Controller('search')
export class SearchController {
  
  @Post('/unified')
  @HttpCode(HttpStatus.OK)
  async unifiedSearch(@Body() searchDto: UnifiedSearchDto) {
    console.log('--- Nhận Request Tìm Kiếm Đã Chuẩn Hóa ---', searchDto);

    // Giả lập logic điều phối (Orchestrator) dựa trên Contract
    if (searchDto.query && searchDto.location) {
      return {
        type: 'HYBRID_SEARCH',
        message: `finding '${searchDto.query}' within ${searchDto.radius_in_meters}m of your location.`,
        data: []
      };
    }

    if (searchDto.location) {
      return {
        type: 'GEO_SEARCH',
        message: `finding locations near: Lat ${searchDto.location.lat}, Lng ${searchDto.location.lng}`,
        data: []
      };
    }

    if (searchDto.query) {
      return {
        type: 'VECTOR_SEARCH',
        message: `finding products matching: '${searchDto.query}'`,
        data: []
      };
    }

    return {
      type: 'NORMAL_SEARCH',
      message: 'No special parameters provided, returning default list.',
      data: []
    };
  }
}