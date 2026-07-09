import { Injectable, Logger } from '@nestjs/common';
import { BaseConsumer } from './base.consumer';
import { RABBITMQ } from '../rabbitmq/rabbitmq.constants';
import { MailService } from '../../modules/mail/mail.service';
import { Event } from '../types/event.interface';
import { IsEmail, IsNotEmpty, IsString } from 'class-validator';

export class UserRegisteredPayload {
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
export class UserRegisteredConsumer extends BaseConsumer<UserRegisteredPayload> {
  protected readonly logger = new Logger(UserRegisteredConsumer.name);
  protected readonly queueName = RABBITMQ.QUEUES.NOTIFICATION; // legacy — no longer active
  protected readonly payloadClass = UserRegisteredPayload;

  constructor(private readonly emailService: MailService) {
    super();
  }

  /**
   * Business logic implementation for user.registered event.
   * Calls the email service to send an OTP code.
   */
  protected async processPayload(payload: UserRegisteredPayload, event: Event<UserRegisteredPayload>): Promise<void> {
    this.logger.log(`Processing user registration event for email: ${payload.email}`);
    await this.emailService.sendOTP(payload.email, payload.full_name, payload.otp_code);
  }
}
