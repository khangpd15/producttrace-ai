import { IsNumber, IsNotEmpty, IsOptional, IsString, Min, Max } from 'class-validator';
import { Type } from 'class-transformer';

export class SearchGeoDto {
  @Type(() => Number)
  @IsNumber()
  @IsNotEmpty()
  @Min(-90)
  @Max(90)
  lat!: number;

  @Type(() => Number)
  @IsNumber()
  @IsNotEmpty()
  @Min(-180)
  @Max(180)
  lng!: number;

  @Type(() => Number)
  @IsNumber()
  @IsNotEmpty()
  @IsOptional()
  radius!: number;

  @IsString()
  @IsNotEmpty()
  @IsOptional() 
  keyword?: string; // (ví dụ: iPhone 15 Pro)
}