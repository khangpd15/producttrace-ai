import { IsNumber, IsOptional, IsString, Min, Max } from 'class-validator';

export class SearchGeoDto {
  @IsNumber()
  lat!: number; // Vĩ độ

  @IsNumber()
  lng!: number; // Kinh độ

  @IsOptional()
  @IsNumber()
  @Min(0)
  radius?: number; // Bán kính tìm kiếm (km)

  @IsOptional()
  @IsString()
  keyword?: string; // Từ khóa tìm kiếm thêm (nếu có)
}