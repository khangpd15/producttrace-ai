// src/modules/search/dto/unified-search.dto.ts
import { IsString, IsOptional, IsNumber, ValidateNested, Min, Max } from 'class-validator';
import { Type } from 'class-transformer';

//Định nghĩa cấu trúc Tọa độ chuẩn (Dùng cho cả Tìm cửa hàng lẫn Tìm bán kính)
class LocationDto {
  @IsNumber()
  @Min(-90)
  @Max(90)
  lat!: number;

  @IsNumber()
  @Min(-180)
  @Max(180)
  lng!: number;
}

// Gộp chung tất cả các kiểu tìm kiếm
export class UnifiedSearchDto {
  @IsOptional()
  @IsString()
  query?: string; // Dùng cho Vector Search (Ví dụ: "áo khoác nam")

  @IsOptional()
  @ValidateNested()
  @Type(() => LocationDto)
  location?: LocationDto; // Dùng cho Geo Search (Vĩ độ & Kinh độ của người dùng)

  @IsOptional()
  @IsNumber()
  @Min(0)
  radius_in_meters?: number = 5000; // Dùng cho Radius Search (Mặc định bán kính 5km)

  @IsOptional()
  @IsNumber()
  @Min(1)
  limit?: number = 10; // Giới hạn số lượng kết quả trả về
}