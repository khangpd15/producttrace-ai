import { IsNumber, IsOptional, Min, Max } from 'class-validator';
import { Type } from 'class-transformer';

export class SearchServiceCenterDto {
  @Type(() => Number)
  @IsNumber()
  @Min(-90) @Max(90) // Tọa độ vĩ độ (Latitude) chỉ từ -90 đến 90
  lat!: number;

  @Type(() => Number)
  @IsNumber()
  @Min(-180) @Max(180) // Tọa độ kinh độ (Longitude) chỉ từ -180 đến 180
  lng!: number;

  @Type(() => Number)
  @IsOptional()
  @Min(0) // Bán kính không được âm
  radius?: number = 20000;
}