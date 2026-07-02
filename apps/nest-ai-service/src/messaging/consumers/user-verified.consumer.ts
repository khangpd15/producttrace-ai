import { Injectable, Logger } from '@nestjs/common';
import { BaseConsumer } from './base.consumer';
import { RABBITMQ } from '../rabbitmq/rabbitmq.constants';
import { MailService } from '../../modules/mail/mail.service';
import { Event } from '../types/event.interface';
import { IsEmail, IsNotEmpty, IsString } from 'class-validator';

export class UserVerifiedPayload {
  @IsEmail({}, { message: 'email must be a valid email address' })
  @IsNotEmpty({ message: 'email is required' })
  email!: string;

  @IsString({ message: 'full_name must be a string' })
  @IsNotEmpty({ message: 'full_name is required' })
  full_name!: string;
}

@Injectable()
export class UserVerifiedConsumer extends BaseConsumer<UserVerifiedPayload> {
  protected readonly logger = new Logger(UserVerifiedConsumer.name);
  protected readonly queueName = RABBITMQ.QUEUES.USER_VERIFIED;
  protected readonly payloadClass = UserVerifiedPayload;

  constructor(private readonly emailService: MailService) {
    super();
  }

  /**
   * Business logic implementation for user.verified event.
   * Sends the account activation / verification success email.
   */
  protected async processPayload(payload: UserVerifiedPayload, event: Event<UserVerifiedPayload>): Promise<void> {
    this.logger.log(`Processing user verification success event for email: ${payload.email}`);
    await this.emailService.sendVerificationSuccess(payload.email, payload.full_name);
  }
}
