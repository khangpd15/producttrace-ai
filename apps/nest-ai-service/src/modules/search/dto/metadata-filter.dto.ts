import { IsOptional, IsString } from 'class-validator';

export class MetadataFilterDto {
  @IsString()
  query!: string;

  @IsOptional()
  @IsString()
  category?: string;

  @IsOptional()
  @IsString()
  manufacturer?: string;

  @IsOptional()
  @IsString()
  province?: string;
}