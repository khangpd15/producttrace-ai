import { Body, Controller, Post } from '@nestjs/common';
import { SearchService } from './search.service';
import { VectorSearchDto } from './dto/vector-search.dto';

@Controller('search')
export class SearchController {
  constructor(
    private readonly searchService: SearchService,
  ) { }

  @Post('vector')
  async vectorSearch(
    @Body() dto: VectorSearchDto,
  ) {
    const result = await this.searchService.vectorSearch(dto);

    return {
      success: true,
      message: 'Semantic search successfully.',
      data: result,
    };
  }
}