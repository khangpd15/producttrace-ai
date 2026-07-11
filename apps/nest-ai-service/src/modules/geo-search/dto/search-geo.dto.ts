import { Type } from 'class-transformer';
import { IsNumber, IsOptional, IsPositive, IsString, Max, Min } from 'class-validator';

export class SearchGeoDto {
  @Type(() => Number)
  @IsNumber()
  lat!: number;

  @Type(() => Number)
  @IsNumber()
  lng!: number;

  @IsOptional()
  @Type(() => Number)
  @IsNumber()
  @IsPositive()
  radius?: number; // Bán kính tính bằng mét (mặc định sẽ xử lý ở service)

  @IsOptional()
  @IsString()
  productId?: string; // Dùng khi tìm sản phẩm cụ thể
}