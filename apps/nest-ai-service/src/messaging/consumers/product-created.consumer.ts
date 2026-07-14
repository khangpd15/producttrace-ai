import { Injectable, Logger } from '@nestjs/common';
import { BaseConsumer } from './base.consumer';
import { RABBITMQ } from '../rabbitmq/rabbitmq.constants';
import { Event } from '../types/event.interface';
import { IsNotEmpty, IsString } from 'class-validator';

export class ProductCreatedPayload {
  @IsString({ message: 'product_name must be a string' })
  @IsNotEmpty({ message: 'product_name is required' })
  product_name!: string;

  @IsString({ message: 'manufacturer must be a string' })
  @IsNotEmpty({ message: 'manufacturer is required' })
  manufacturer!: string;

  @IsString({ message: 'batch_code must be a string' })
  @IsNotEmpty({ message: 'batch_code is required' })
  batch_code!: string;
}

@Injectable()
export class ProductCreatedConsumer extends BaseConsumer<ProductCreatedPayload> {
  protected readonly logger = new Logger(ProductCreatedConsumer.name);
  protected readonly queueName = RABBITMQ.QUEUES.NOTIFICATION; // legacy — no longer active
  protected readonly payloadClass = ProductCreatedPayload;

  /**
   * Business logic implementation for product.created event.
   * Currently just logs product trace details to the console.
   */
  protected async processPayload(payload: ProductCreatedPayload, event: Event<ProductCreatedPayload>): Promise<void> {
    this.logger.log(
      `[Product Created Event Processed] Name: ${payload.product_name} | Manufacturer: ${payload.manufacturer} | Batch Code: ${payload.batch_code}`
    );
  }
}