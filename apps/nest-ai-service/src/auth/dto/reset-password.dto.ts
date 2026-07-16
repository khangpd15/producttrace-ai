import { IsNotEmpty, IsString, MinLength } from 'class-validator';

export class ResetPasswordDto {
  @IsNotEmpty()
  @IsString()
  email!: string;

  @IsNotEmpty()
  @IsString()
  otp_code!: string;

  @IsNotEmpty()
  @IsString()
  new_password!: string;
}
