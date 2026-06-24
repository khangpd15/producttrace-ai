import { Controller, Logger } from '@nestjs/common';
import { Ctx, EventPattern, Payload, RmqContext } from '@nestjs/microservices';

import { Event } from '../types/event.interface';
import { RABBITMQ } from '../../integrations/rabbitmq/rabbitmq.constants';

@Controller()
export class ProductCreatedConsumer {
  private readonly logger = new Logger(ProductCreatedConsumer.name);

  @EventPattern(RABBITMQ.ROUTING_KEYS.PRODUCT_CREATED)
  async handle(
    @Payload() event: Event,
    @Ctx() context: RmqContext,
  ) {
    const channel = context.getChannelRef();
    const message = context.getMessage();

    try {
      this.logger.log(JSON.stringify(event));

      channel.ack(message);
    } catch (error) {
      channel.nack(message, false, false);
    }
  }
}