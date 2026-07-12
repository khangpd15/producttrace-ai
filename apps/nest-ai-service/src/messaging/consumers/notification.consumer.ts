import { Injectable, Logger } from '@nestjs/common';
import { IsEmail, IsNotEmpty, IsOptional, IsString } from 'class-validator';
import { BaseConsumer } from './base.consumer';
import { RABBITMQ } from '../rabbitmq/rabbitmq.constants';
import { MailService } from '../../modules/mail/mail.service';
import { Event } from '../types/event.interface';

/**
 * NotificationPayload covers all possible fields across every event type.
 * Fields are optional so that validation does not reject events that only
 * carry a subset of fields (e.g. otp.verified has no otp_code).
 */
export class NotificationPayload {
  @IsEmail({}, { message: 'email must be a valid email address' })
  @IsNotEmpty({ message: 'email is required' })
  email!: string;

  @IsOptional()
  @IsString()
  full_name?: string;

  @IsOptional()
  @IsString()
  otp_code?: string;

  @IsOptional()
  @IsString()
  phone?: string;
}

@Injectable()
export class NotificationConsumer extends BaseConsumer<NotificationPayload> {
  protected readonly logger = new Logger(NotificationConsumer.name);

  /** Consumes the single unified ai.events queue. */
  protected readonly queueName = RABBITMQ.QUEUES.NOTIFICATION;

  protected readonly payloadClass = NotificationPayload;

  constructor(private readonly mailService: MailService) {
    super();
  }

  /**
   * Routes to the correct mail action based on the event_type in the envelope.
   *
   * Event types published by Go Core Service:
   *  "otp.registered"  — Go OTP worker sent registration OTP → sendOTP
   *  "otp.forgot"      — Go ForgotPassword flow            → sendPasswordReset
   *  "otp.verified"    — Go VerifyOTP flow                 → sendVerificationSuccess
   *  "product.created" — Go product service                → (placeholder log)
   */
  protected async processPayload(
    payload: NotificationPayload,
    event: Event<NotificationPayload>,
  ): Promise<void> {
    const eventType = event.event_type;

    this.logger.log(
      `[NotificationConsumer] Routing event_type="${eventType}" email="${payload.email}"`,
    );

    switch (eventType) {
      // ── Registration OTP ──────────────────────────────────────────────────
      case RABBITMQ.EVENT_TYPES.USER_REGISTERED: // "otp.registered"
        if (!payload.otp_code) {
          throw new Error(`Missing otp_code for event_type="${eventType}"`);
        }
        await this.mailService.sendOTP(
          payload.email,
          payload.full_name || 'User',
          payload.otp_code,
        );
        this.logger.log(`[NotificationConsumer] Registration OTP sent to ${payload.email}`);
        break;

      // ── Forgot-Password OTP ───────────────────────────────────────────────
      case RABBITMQ.EVENT_TYPES.PASSWORD_RESET: // "otp.forgot"
        if (!payload.otp_code) {
          throw new Error(`Missing otp_code for event_type="${eventType}"`);
        }
        await this.mailService.sendPasswordReset(
          payload.email,
          payload.full_name || 'User',
          payload.otp_code,
        );
        this.logger.log(`[NotificationConsumer] Password-reset OTP sent to ${payload.email}`);
        break;

      // ── Account Verified ──────────────────────────────────────────────────
      case RABBITMQ.EVENT_TYPES.USER_VERIFIED: // "otp.verified"
        await this.mailService.sendVerificationSuccess(
          payload.email,
          payload.full_name || 'User',
        );
        this.logger.log(`[NotificationConsumer] Verification-success email sent to ${payload.email}`);
        break;

      // ── Unknown / unhandled ───────────────────────────────────────────────
      default:
        this.logger.warn(
          `[NotificationConsumer] Unhandled event_type="${eventType}" — message acknowledged without action.`,
        );
    }
  }
}
