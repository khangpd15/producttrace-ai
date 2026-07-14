import { Event } from '../messaging/types/event.interface';

export class BuildTextUtil {
  static buildText(event: Event): string {
    const eventType = (event.event_type ?? '').toLowerCase();
    if (eventType.includes('trace')) {
      return this.buildTraceText(event.payload as Record<string, unknown>);
    }

    if (eventType.includes('warranty')) {
      return this.buildWarrantyText(event.payload as Record<string, unknown>);
    }

    return this.buildProductText(event.payload as Record<string, unknown>);
  }

  private static buildProductText(payload: Record<string, unknown>): string {
    const name = String(payload.name ?? payload.productName ?? '');
    const brand = String(payload.brand ?? '');
    const category = String(payload.category ?? '');
    const description = String(payload.description ?? payload.productDescription ?? '');
    const location = String(payload.location ?? '');
    const tags = this.arrayToString(payload.tags);
    const attributes = this.attributesToString(payload.attributes);

    return [
      `Product: ${name}`,
      `Brand: ${brand}`,
      `Category: ${category}`,
      `Location: ${location}`,
      `Description: ${description}`,
      `Tags: ${tags}`,
      `Attributes: ${attributes}`,
    ].join('\n');
  }

  private static buildTraceText(payload: Record<string, unknown>): string {
    const batchId = String(payload.batchId ?? payload.batch ?? '');
    const productId = String(payload.productId ?? payload.product ?? '');
    const status = String(payload.status ?? '');
    const location = String(payload.location ?? '');
    const timestamp = String(payload.timestamp ?? payload.time ?? '');

    return [
      'Trace Event:',
      `Batch: ${batchId}`,
      `Product: ${productId}`,
      `Status: ${status}`,
      `Location: ${location}`,
      `Time: ${timestamp}`,
    ].join('\n');
  }

  private static buildWarrantyText(payload: Record<string, unknown>): string {
    const productId = String(payload.productId ?? payload.product ?? '');
    const warrantyPeriod = String(payload.warrantyPeriod ?? payload.warranty ?? '');
    const provider = String(payload.provider ?? '');
    const status = String(payload.status ?? '');

    return [
      `Product: ${productId}`,
      `Warranty: ${warrantyPeriod}`,
      `Provider: ${provider}`,
      `Status: ${status}`,
    ].join('\n');
  }

  private static arrayToString(input: unknown): string {
    if (Array.isArray(input)) {
      return input.map((item) => String(item)).join(', ');
    }
    if (typeof input === 'string') {
      return input;
    }
    return '';
  }

  private static attributesToString(input: unknown): string {
    if (typeof input === 'object' && input !== null && !Array.isArray(input)) {
      return Object.entries(input)
        .map(([key, value]) => `${key}: ${this.valueToString(value)}`)
        .join(', ');
    }

    return this.arrayToString(input);
  }

  private static valueToString(value: unknown): string {
    if (Array.isArray(value)) {
      return value.map((item) => String(item)).join(', ');
    }
    return String(value ?? '');
  }
}
