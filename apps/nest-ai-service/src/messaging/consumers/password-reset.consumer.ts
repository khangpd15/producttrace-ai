import { Controller, Logger } from '@nestjs/common';
import { Ctx, EventPattern, Payload, RmqContext } from '@nestjs/microservices';

import { RABBITMQ } from '../rabbitmq/rabbitmq.constants';
import { MailService } from '../../modules/mail/mail.service';
import { ConfigService } from '@nestjs/config';

export interface PasswordResetRequestedEvent {
  email: string;
  name: string;
  resetToken: string;
  resetLink?: string;
  [key: string]: any;
}

@Controller()
export class PasswordResetConsumer {
  private readonly logger = new Logger(PasswordResetConsumer.name);

  constructor(
    private readonly mailService: MailService,
    private readonly configService: ConfigService,
  ) {}

  @EventPattern(RABBITMQ.ROUTING_KEYS.PASSWORD_RESET_REQUESTED)
  async handle(
    @Payload() event: PasswordResetRequestedEvent,
    @Ctx() context: RmqContext,
  ) {
    const channel = context.getChannelRef();
    const message = context.getMessage();

    try {
      this.logger.log(`Received auth.password_reset_requested event for: ${event.email}`);
      
      const templateId = this.configService.get<string>('RESET_PASSWORD_TEMPLATE_ID') || '';
      
      // Send the forgot password template email
      await this.mailService.sendTemplateMail(event.email, templateId, {
        name: event.name,
        resetToken: event.resetToken,
        resetLink: event.resetLink,
      });

      // Acknowledge message processing is successful
      channel.ack(message);
    } catch (error) {
      this.logger.error('Error processing auth.password_reset_requested event', error);
      // Negative acknowledgement
      channel.nack(message, false, false);
    }
  }
}
