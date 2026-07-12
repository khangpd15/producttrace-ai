import { Body, Controller, Post } from '@nestjs/common';

import { SearchService } from './search.service';
import { HybridSearchDto } from './dto/hybrid-search.dto';

@Controller('search')
export class SearchController {
  constructor(
    private readonly searchService: SearchService,
  ) {}

  @Post('hybrid')
  async hybridSearch(
    @Body() dto: HybridSearchDto,
  ) {
    return this.searchService.hybridSearch(dto);
  }
}