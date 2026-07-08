import { IsNotEmpty, IsOptional, IsString, IsInt, Min, Max } from 'class-validator';
import { SearchFilter } from '../interfaces/search-filter.interface';

export class VectorSearchDto {
  @IsString()
  @IsNotEmpty()
  query!: string;

  @IsOptional()
  filter?: SearchFilter;
  @IsInt()
  @Min(1)
  @Max(50)
  limit?: number = 10;
}