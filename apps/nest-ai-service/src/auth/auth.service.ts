import { Injectable } from '@nestjs/common';
import * as crypto from 'crypto';
import * as bcrypt from 'bcrypt';
import { MailService } from '../mail/mail.service';
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

    // Generate secure token
    const token = crypto.randomBytes(32).toString('hex');
    
    // Hash token for secure storage
    const hashedToken = crypto.createHash('sha256').update(token).digest('hex');
    
    // Expiration = 15 minutes
    const expiresAt = new Date();
    expiresAt.setMinutes(expiresAt.getMinutes() + 15);

    // Save hashed token
    await this.passwordResetRepository.saveToken(email, hashedToken, expiresAt);

    // Build reset URL
    const frontendUrl = this.configService.get<string>('FRONTEND_URL') || 'http://localhost:5173';
    const resetUrl = `${frontendUrl}/reset-password?token=${token}&email=${encodeURIComponent(email)}`;

    // Send email
    await this.mailService.sendPasswordResetEmail(email, resetUrl, user.name);

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

  async resetPassword(email: string, resetPasswordDto: ResetPasswordDto): Promise<{ message: string }> {
    const { token, password } = resetPasswordDto;

    const hashedToken = crypto.createHash('sha256').update(token).digest('hex');
    const record = await this.passwordResetRepository.findByTokenAndEmail(hashedToken, email);

    if (!record) {
      throw new Error('Invalid or already used reset link.');
    }

    if (new Date(record.expiresAt) < new Date()) {
      throw new Error('This reset link has expired.');
    }

    const user = await this.userRepository.findByEmail(email);
    if (!user) {
      throw new Error('User not found.');
    }

    // Hash new password using bcrypt (12 rounds)
    const newHashedPassword = await bcrypt.hash(password, 12);

    // Update password
    await this.userRepository.updatePassword(user.id, newHashedPassword);

    // Remove used token (one-time token)
    await this.passwordResetRepository.deleteTokenByEmail(email);

    return { message: 'Password changed successfully.' };
  }
}
