import { Injectable, Logger } from '@nestjs/common';
import { BaseConsumer } from './base.consumer';
import { RABBITMQ } from '../rabbitmq/rabbitmq.constants';
import { MailService } from '../../modules/mail/mail.service';
import { Event } from '../types/event.interface';
import { IsEmail, IsNotEmpty, IsString } from 'class-validator';

export class OwnershipPayload {
    @IsEmail({}, { message: 'email must be a valid email address' })
    @IsNotEmpty({ message: 'email is required' })
    email!: string;

    @IsString({ message: 'otp_code must be a string' })
    @IsNotEmpty({ message: 'otp_code is required' })
    otp_code!: string;

    @IsString({ message: 'product_id must be a string' })
    @IsNotEmpty({ message: 'product_id is required' })
    product_id!: string;
}

@Injectable()
export class OwnershipConsumer extends BaseConsumer<OwnershipPayload> {
    protected readonly logger = new Logger(OwnershipConsumer.name);
    protected readonly queueName = RABBITMQ.QUEUES.NOTIFICATION;
    protected readonly payloadClass = OwnershipPayload;

    constructor(private readonly emailService: MailService) {
        super();
    }

    /**
     * Business logic implementation for user.registered event.
     * Calls the email service to send an OTP code.
     */
    protected async processPayload(payload: OwnershipPayload, event: Event<OwnershipPayload>): Promise<void> {
        this.logger.log(`Processing ownership event for email: ${payload.email}`);
        await this.emailService.sendOTP(payload.email, payload.otp_code, payload.product_id);
    }
}
