import { IsOptional, IsString } from 'class-validator';

export class FilterDto {

    @IsOptional()
    @IsString()
    brand?: string;

    @IsOptional()
    @IsString()
    category?: string;

}