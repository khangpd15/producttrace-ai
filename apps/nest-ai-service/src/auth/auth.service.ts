import { Injectable } from '@nestjs/common';
import * as crypto from 'crypto';
import * as bcrypt from 'bcrypt';
import { MailService } from '../modules/mail/mail.service';
import { ForgotPasswordDto } from './dto/forgot-password.dto';
import { ResetPasswordDto } from './dto/reset-password.dto';
import { Inject } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { USER_REPOSITORY, UserRepository } from './repositories/user.repository';
import { PASSWORD_RESET_REPOSITORY, PasswordResetRepository } from './repositories/password-reset.repository';

@Injectable()
export class AuthService {
  constructor(
    @Inject(USER_REPOSITORY) private readonly userRepository: UserRepository,
    @Inject(PASSWORD_RESET_REPOSITORY) private readonly passwordResetRepository: PasswordResetRepository,
    private readonly mailService: MailService,
    private readonly configService: ConfigService,
  ) {}

  async forgotPassword(forgotPasswordDto: ForgotPasswordDto): Promise<{ message: string }> {
    const { email } = forgotPasswordDto;
    
    // Always return the same response to prevent email enumeration
    const successMessage = { message: 'If an account with that email exists, a password reset email has been sent.' };

    const user = await this.userRepository.findByEmail(email);
    if (!user) {
      return successMessage;
    }

    // Generate secure 6-digit OTP code
    const otp = Math.floor(100000 + Math.random() * 900000).toString();
    
    // Hash OTP for secure storage
    const hashedToken = crypto.createHash('sha256').update(otp).digest('hex');
    
    // Expiration = 15 minutes
    const expiresAt = new Date();
    expiresAt.setMinutes(expiresAt.getMinutes() + 15);

    // Save hashed token
    await this.passwordResetRepository.saveToken(email, hashedToken, expiresAt);

    // Send OTP email
    await this.mailService.sendPasswordReset(email, user.name, otp);

    return successMessage;
  }

  async validateResetToken(token: string, email: string): Promise<{ valid: boolean; message: string }> {
    const hashedToken = crypto.createHash('sha256').update(token).digest('hex');
    const record = await this.passwordResetRepository.findByTokenAndEmail(hashedToken, email);

    if (!record) {
      return { valid: false, message: 'Invalid or already used reset link.' };
    }

    if (new Date(record.expiresAt) < new Date()) {
      return { valid: false, message: 'This reset link has expired.' };
    }

    return { valid: true, message: 'Token is valid.' };
  }

  async resetPassword(resetPasswordDto: ResetPasswordDto): Promise<{ message: string }> {
    const { email, otp_code, new_password } = resetPasswordDto;

    const hashedToken = crypto.createHash('sha256').update(otp_code).digest('hex');
    const record = await this.passwordResetRepository.findByTokenAndEmail(hashedToken, email);

    if (!record) {
      throw new Error('Mã OTP không hợp lệ hoặc đã sử dụng.');
    }

    if (new Date(record.expiresAt) < new Date()) {
      throw new Error('Mã OTP đã hết hạn.');
    }

    const user = await this.userRepository.findByEmail(email);
    if (!user) {
      throw new Error('Người dùng không tồn tại.');
    }

    // Hash new password using bcrypt (12 rounds)
    const newHashedPassword = await bcrypt.hash(new_password, 12);

    // Update password
    await this.userRepository.updatePassword(user.id, newHashedPassword);

    // Remove used token (one-time token)
    await this.passwordResetRepository.deleteTokenByEmail(email);

    return { message: 'Password changed successfully.' };
  }
}
