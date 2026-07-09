import { Logger } from '@nestjs/common';
import * as amqp from 'amqplib';
import { Event } from '../types/event.interface';
import { validate } from 'class-validator';
import { plainToInstance } from 'class-transformer';

/**
 * BaseConsumer provides a standard structure for consuming events from RabbitMQ.
 * It handles JSON parsing, payload validation, execution retry (3 times),
 * and standard event logging (Received event, Event Type, Correlation ID, Email, Success, Failure).
 */
export abstract class BaseConsumer<T> {
  protected abstract readonly logger: Logger;
  protected abstract readonly queueName: string;
  protected abstract readonly payloadClass: new (...args: any[]) => T;

  /**
   * Processes the validated payload.
   * If this throws an error, the message will be nacked and not requeued (sent to DLX/discarded).
   */
  protected abstract processPayload(payload: T, event: Event<T>): Promise<void>;

  /**
   * Handles the raw RabbitMQ message, parses it, validates it, and logs the execution.
   */
  async handleMessage(msg: amqp.ConsumeMessage, channel: amqp.Channel): Promise<void> {
    const rawContent = msg.content.toString();
    let parsedEvent: Event<any> | null = null;
    let eventType = 'unknown';
    let correlationId = 'unknown';
    let email = 'N/A';

    try {
      // 1. Parse JSON event
      try {
        const parsed = JSON.parse(rawContent);
        if (parsed && typeof parsed === 'object' && 'pattern' in parsed && 'data' in parsed) {
          parsedEvent = parsed.data as Event<any>;
        } else {
          parsedEvent = parsed as Event<any>;
        }
      } catch (parseErr) {
        throw new Error(`Failed to parse message JSON: ${rawContent}`);
      }

      eventType = parsedEvent?.event_type || 'unknown';
      correlationId = parsedEvent?.correlation_id || 'unknown';

      // Safe extract of email if it exists in the payload
      if (parsedEvent?.payload && typeof parsedEvent.payload === 'object') {
        email = parsedEvent.payload.email || parsedEvent.payload.email_address || 'N/A';
      }

      // Log receipt
      this.logger.log(
        `[Received event] Event Type: ${eventType} | Correlation ID: ${correlationId} | Email: ${email} | Queue: ${this.queueName}`
      );

      // 2. Validate basic structure
      if (!parsedEvent || !parsedEvent.event_id || !parsedEvent.event_type || parsedEvent.payload === undefined) {
        throw new Error('Invalid event structure: missing event_id, event_type, or payload');
      }

      // 3. Class validation of the payload
      const payloadInstance = plainToInstance(this.payloadClass, parsedEvent.payload) as any;
      const errors = await validate(payloadInstance);

      if (errors.length > 0) {
        const errorMsg = errors
          .map((err: any) => Object.values(err.constraints || {}).join(', '))
          .join('; ');
        throw new Error(`Payload validation failed: ${errorMsg}`);
      }

      // 4. Process payload with a light retry mechanism (3 attempts)
      await this.processWithRetry(payloadInstance, parsedEvent);

      // 5. Acknowledge message upon successful processing
      channel.ack(msg);
      this.logger.log(
        `[Success] Event Type: ${eventType} | Correlation ID: ${correlationId} | Email: ${email} | Acknowledged.`
      );
    } catch (error: any) {
      this.logger.error(
        `[Failure] Event Type: ${eventType} | Correlation ID: ${correlationId} | Email: ${email} | Error: ${error.message}`
      );

      // 6. Reject message (requeue = false) to send to DLQ or discard it
      try {
        channel.nack(msg, false, false);
      } catch (nackError) {
        this.logger.error('Failed to nack message after failure', nackError);
      }
    }
  }

  /**
   * Helper to perform a light retry when processing fails (e.g. sending email failure).
   */
  private async processWithRetry(payload: T, event: Event<T>): Promise<void> {
    const maxRetries = 3;
    let attempt = 0;

    while (attempt < maxRetries) {
      try {
        attempt++;
        await this.processPayload(payload, event);
        return; // Success, exit retry loop
      } catch (error: any) {
        this.logger.warn(
          `Processing attempt ${attempt}/${maxRetries} failed for Event ${event.event_id}: ${error.message}`
        );
        if (attempt >= maxRetries) {
          throw error; // Re-throw to trigger reject/nack
        }
        // Small delay (1s) before retrying
        await new Promise((resolve) => setTimeout(resolve, 1000));
      }
    }
  }
}
