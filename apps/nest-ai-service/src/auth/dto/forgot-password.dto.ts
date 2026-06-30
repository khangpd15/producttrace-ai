import { IsEmail, IsNotEmpty } from 'class-validator';

export class ForgotPasswordDto {
  @IsNotEmpty()
  @IsEmail({}, { message: 'Please provide a valid email address' })
  email!: string;
}
