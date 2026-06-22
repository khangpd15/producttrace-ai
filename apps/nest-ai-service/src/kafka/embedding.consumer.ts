import { Controller, Logger } from '@nestjs/common';
import { Ctx, EventPattern, Payload, KafkaContext } from '@nestjs/microservices';

import { EmbeddingService } from '../modules/embedding/embedding.service';
import { KAFKA } from './kafka.constants';

@Controller()
export class EmbeddingConsumer {
  private readonly logger = new Logger(EmbeddingConsumer.name);

  constructor(private readonly embeddingService: EmbeddingService) { }

  @EventPattern(KAFKA.TOPICS.PRODUCT_EVENTS)
  async consumeProductEvent(@Payload() payload: unknown, @Ctx() context: KafkaContext) {
    return this.handleEvent(payload, context);
  }

  @EventPattern(KAFKA.TOPICS.TRACE_EVENTS)
  async consumeTraceEvent(@Payload() payload: unknown, @Ctx() context: KafkaContext) {
    return this.handleEvent(payload, context);
  }

  private async handleEvent(payload: unknown, context: KafkaContext) {
    const event = this.normalizeEvent(payload);
    const topic = context.getTopic();
    const partition = context.getPartition();
    const message = context.getMessage();
    const offset = message?.offset ?? 'unknown';

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

    this.logger.log(`Received kafka event topic=${topic} partition=${partition} offset=${offset} id=${event.eventId} type=${event.eventType}`);

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

    // KafkaJS message wrapper
    if (data?.value) {
      data = data.value;
    }

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
