import { Controller, Post, Body, Get, Query, HttpException, HttpStatus, BadRequestException } from '@nestjs/common';
import { AuthService } from './auth.service';
import { ForgotPasswordDto } from './dto/forgot-password.dto';
import { ResetPasswordDto } from './dto/reset-password.dto';

@Controller('auth')
export class AuthController {
  constructor(private readonly authService: AuthService) {}

  @Post('forgot-password')
  async forgotPassword(@Body() forgotPasswordDto: ForgotPasswordDto) {
    try {
      return await this.authService.forgotPassword(forgotPasswordDto);
    } catch (error: any) {
      // Centralized exception handling
      throw new HttpException(error.message || 'Internal server error', HttpStatus.INTERNAL_SERVER_ERROR);
    }
  }

  @Get('validate-reset-token')
  async validateResetToken(@Query('token') token: string, @Query('email') email: string) {
    if (!token || !email) {
      throw new BadRequestException('Token and email are required');
    }
    const result = await this.authService.validateResetToken(token, email);
    if (!result.valid) {
      throw new HttpException(result.message, HttpStatus.BAD_REQUEST);
    }
    return result;
  }

  @Post('reset-password')
  async resetPassword(@Body() resetPasswordDto: ResetPasswordDto) {
    try {
      return await this.authService.resetPassword(resetPasswordDto);
    } catch (error: any) {
      throw new HttpException(error.message || 'Bad Request', HttpStatus.BAD_REQUEST);
    }
  }
}
