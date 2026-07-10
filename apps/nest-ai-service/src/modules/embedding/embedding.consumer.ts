import { Controller, Logger } from '@nestjs/common';
import { EventPattern, Payload, Ctx, RmqContext } from '@nestjs/microservices';
import { EmbeddingService } from './embedding.service';
import { RABBITMQ } from '../../integrations/rabbitmq/rabbitmq.constants';

@Controller()
export class EmbeddingConsumer {
  private readonly logger = new Logger(EmbeddingConsumer.name);

  constructor(private readonly embeddingService: EmbeddingService) { }

  @EventPattern(RABBITMQ.ROUTING_KEYS.PRODUCT_CREATED)
  async consumeProductEvent(@Payload() payload: unknown, @Ctx() context: RmqContext) {
    return this.handleEvent(payload, context);
  }

  @EventPattern(RABBITMQ.ROUTING_KEYS.TRACE_EVENTS)
  async consumeTraceEvent(@Payload() payload: unknown, @Ctx() context: RmqContext) {
    return this.handleEvent(payload, context);
  }

  private async handleEvent(payload: unknown, context: RmqContext) {
    const event = this.normalizeEvent(payload);
    const message = context.getMessage();

    this.logger.debug(
      `EVENT PAYLOAD = ${JSON.stringify(event)}`
    );

    if (!event) {
      this.logger.error('Empty event received');
      return;
    }

    if (!event.eventType) {
      this.logger.error(`Invalid event format: ${JSON.stringify(event)}`);
      return;
    }

    this.logger.log(
      `Received rabbitmq event id=${event.eventId} type=${event.eventType}`
    );

    try {
      await this.embeddingService.processEvent(event);
      this.logger.log(`Processed event ${event.eventId} type=${event.eventType}`);
    } catch (error) {
      this.logger.error(`Failed processing event ${event.eventId}: ${error instanceof Error ? error.message : JSON.stringify(error)}`);
      throw error;
    }
  }

  private normalizeEvent(payload: any): any {
    let data = payload;

    // Buffer -> string
    if (Buffer.isBuffer(data)) {
      data = data.toString('utf8');
    }

    // string -> JSON
    if (typeof data === 'string') {
      try {
        data = JSON.parse(data);
      } catch (e) {
        this.logger.error('Invalid JSON payload');
        throw e;
      }
    }

    return data;
  }
}
