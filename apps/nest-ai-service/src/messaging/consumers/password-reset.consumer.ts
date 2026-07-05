import { Injectable, Logger } from '@nestjs/common';
import { BaseConsumer } from './base.consumer';
import { RABBITMQ } from '../rabbitmq/rabbitmq.constants';
import { MailService } from '../../modules/mail/mail.service';
import { Event } from '../types/event.interface';
import { IsEmail, IsNotEmpty, IsString } from 'class-validator';

export class PasswordResetPayload {
  @IsEmail({}, { message: 'email must be a valid email address' })
  @IsNotEmpty({ message: 'email is required' })
  email!: string;

  @IsString({ message: 'full_name must be a string' })
  @IsNotEmpty({ message: 'full_name is required' })
  full_name!: string;

  @IsString({ message: 'otp_code must be a string' })
  @IsNotEmpty({ message: 'otp_code is required' })
  otp_code!: string;
}

@Injectable()
export class PasswordResetConsumer extends BaseConsumer<PasswordResetPayload> {
  protected readonly logger = new Logger(PasswordResetConsumer.name);
  protected readonly queueName = RABBITMQ.QUEUES.PASSWORD_RESET;
  protected readonly payloadClass = PasswordResetPayload;

  constructor(private readonly emailService: MailService) {
    super();
  }

  /**
   * Business logic implementation for password.reset event.
   * Calls the email service to send a password reset OTP.
   */
  protected async processPayload(payload: PasswordResetPayload, event: Event<PasswordResetPayload>): Promise<void> {
    this.logger.log(`Processing password reset event for email: ${payload.email}`);
    await this.emailService.sendPasswordReset(payload.email, payload.full_name, payload.otp_code);
  }
}
