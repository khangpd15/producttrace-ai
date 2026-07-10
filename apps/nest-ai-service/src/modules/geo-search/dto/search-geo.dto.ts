import { Type } from 'class-transformer';
import { IsNumber, IsOptional, IsPositive, IsString, Max, Min } from 'class-validator';

export class SearchGeoDto {
  @Type(() => Number)
  @IsNumber()
  @Min(-90)
  @Max(90)
  lat!: number;

  @Type(() => Number)
  @IsNumber()
  @Min(-180)
  @Max(180)
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