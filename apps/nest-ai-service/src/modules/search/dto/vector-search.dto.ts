import { IsNotEmpty, IsString } from 'class-validator';

export class VectorSearchDto {
  @IsString()
  @IsNotEmpty()
  query!: string;
}