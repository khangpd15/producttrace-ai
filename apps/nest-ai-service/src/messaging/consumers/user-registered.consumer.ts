import { Controller, Logger } from '@nestjs/common';
import { Ctx, EventPattern, Payload, RmqContext } from '@nestjs/microservices';

import { MailService } from '../../modules/mail/mail.service';
import { ConfigService } from '@nestjs/config';
import { RABBITMQ } from '../../integrations/rabbitmq/rabbitmq.constants';

export interface UserRegisteredEvent {
  email: string;
  name: string;
  [key: string]: any;
}

@Controller()
export class UserRegisteredConsumer {
  private readonly logger = new Logger(UserRegisteredConsumer.name);

  constructor(
    private readonly mailService: MailService,
    private readonly configService: ConfigService,
  ) {}

  @EventPattern(RABBITMQ.ROUTING_KEYS.USER_REGISTERED)
  async handle(
    @Payload() event: UserRegisteredEvent,
    @Ctx() context: RmqContext,
  ) {
    const channel = context.getChannelRef();
    const message = context.getMessage();

    try {
      this.logger.log(`Received user.registered event for: ${event.email}`);
      
      const templateId = this.configService.get<string>('WELCOME_TEMPLATE_ID') || '';
      
      // We will send a template email
      // The template may need the user's name
      await this.mailService.sendTemplateMail(event.email, templateId, {
        name: event.name,
      });

      // Acknowledge message processing is successful
      channel.ack(message);
    } catch (error) {
      this.logger.error('Error processing user.registered event', error);
      // Negative acknowledgement - optionally requeue
      channel.nack(message, false, false);
    }
  }
}
