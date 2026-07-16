/// <reference types="jest" />
import { Test, TestingModule } from '@nestjs/testing';
import { GeoSearchController } from './geo-search.controller';
import { QdrantService } from '../../integrations/qdrant/qdrant.service';
import { SearchGeoDto } from './dto/search-geo.dto';

describe('GeoSearchController', () => {
  let controller: GeoSearchController;
  let qdrantService: QdrantService;

  const mockQdrantService = {
    findStoresByRadius: jest.fn(),
  };

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [GeoSearchController],
      providers: [
        {
          provide: QdrantService,
          useValue: mockQdrantService,
        },
      ],
    }).compile();

    controller = module.get<GeoSearchController>(GeoSearchController);
    qdrantService = module.get<QdrantService>(QdrantService);
  });

  it('Should be defined', () => {
    expect(controller).toBeDefined();
  });

  it('Should return a list of nearby stores successfully', async () => {
    const mockStores = [
      { id: 1, name: 'Cửa hàng FPT Cần Thơ', location: { lat: 10.0226, lon: 105.7314 } },
    ];
    mockQdrantService.findStoresByRadius.mockResolvedValue(mockStores);

    // Khai báo đúng định dạng số của DTO chuẩn hóa
    const mockQuery: SearchGeoDto = { lat: 10.0226, lng: 105.7314, radius: 5000 };

    const result = await controller.getNearestStore(mockQuery);

    expect(result).toEqual({
      success: true,
      message: 'Found 1 stores within 5000m',
      data: mockStores,
    });

    expect(mockQdrantService.findStoresByRadius).toHaveBeenCalledWith(10.0226, 105.7314, 5000);
  });
});