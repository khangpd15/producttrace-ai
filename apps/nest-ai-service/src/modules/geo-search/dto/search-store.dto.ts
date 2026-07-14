import { IsNumber, IsNotEmpty, IsOptional, IsString, Min, Max } from 'class-validator';
import { Type } from 'class-transformer';

export class SearchStoreDto {
  @Type(() => Number)
  @IsNumber()
  @Min(-90) @Max(90)
  lat!: number;
  @Type(() => Number)
  @IsNumber()
  @Min(-180) @Max(180)
  lng!: number;
  
  @Type(() => Number)
  @IsOptional()
  @Min(0)
  radius?: number = 20000;

  @IsString()
  @IsOptional() 
  keyword?: string;
}