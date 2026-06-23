import { Controller, Get, Query } from '@nestjs/common';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';

@Controller('geo')
export class GeoSearchController {
  // Inject QdrantService vào để dùng
  constructor(private readonly qdrantService: QdrantService) {}

  @Get('/nearest-store')
  async getNearestStore(
    @Query('lat') lat: string,
    @Query('lng') lng: string,
    @Query('radius') radius?: string,
  ) {
    console.log(`[GET] Request finding nearest stores: Lat ${lat}, Lng ${lng}`);

    const latitude = parseFloat(lat);
    const longitude = parseFloat(lng);
    const radiusMeters = radius ? parseInt(radius) : 5000; // Mặc định 5km nếu không truyền

    // Gọi trực tiếp vào hàm quét tọa độ thật của Qdrant
    const stores = await this.qdrantService.findStoresByRadius(latitude, longitude, radiusMeters);

    return {
      success: true,
      message: `Has find ${stores.length} stores within ${radiusMeters}m`,
      data: stores,
    };
  }
}