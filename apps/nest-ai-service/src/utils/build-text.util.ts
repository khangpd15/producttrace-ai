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
    // Support both flat fields and nested metadata.brand / metadata.category
    const metadata = payload.metadata as Record<string, unknown> | undefined;
    const brand = String(payload.brand ?? metadata?.brand ?? '');
    const category = String(payload.category ?? metadata?.category ?? '');
    const description = String(payload.description ?? payload.productDescription ?? '');
    const location = String(payload.location ?? '');
    const status = String(payload.status ?? '');
    const slug = String(payload.slug ?? '');
    const tags = this.arrayToString(payload.tags);

    // Variants: extract SKU and variant names as additional search keywords
    const variants = payload.variants as Array<Record<string, unknown>> | undefined;
    let variantSkus = '';
    let variantNames = '';
    let variantAttributes = '';
    if (Array.isArray(variants)) {
      const skus: string[] = [];
      const names: string[] = [];
      const attrParts: string[] = [];
      for (const v of variants) {
        if (v.sku) skus.push(String(v.sku));
        if (v.name) names.push(String(v.name));
        // Variant-level attributes (array of { label, value_text, value_number, value_boolean })
        const attrs = v.attributes as Array<Record<string, unknown>> | undefined;
        if (Array.isArray(attrs)) {
          for (const a of attrs) {
            const label = String(a.label ?? '');
            const val = String(a.value_text ?? a.value_number ?? a.value_boolean ?? '');
            if (label && val) attrParts.push(`${label}: ${val}`);
          }
        }
      }
      variantSkus = skus.join(', ');
      variantNames = names.join(', ');
      variantAttributes = attrParts.join(', ');
    }

    // Attributes from payload root (object format, e.g. { color: "Black", storage: "256GB" })
    const rootAttributes = this.attributesToString(payload.attributes);

    // Combine variant attributes with root-level attributes
    const allAttributes = [rootAttributes, variantAttributes]
      .filter(part => part.length > 0)
      .join(', ');

    const lines: string[] = [
      `Product: ${name}`,
    ];

    if (brand) lines.push(`Brand: ${brand}`);
    if (category) lines.push(`Category: ${category}`);
    if (slug) lines.push(`Slug: ${slug}`);
    if (status) lines.push(`Status: ${status}`);
    if (description) lines.push(`Description: ${description}`);
    if (location) lines.push(`Location: ${location}`);
    if (tags) lines.push(`Tags: ${tags}`);
    if (variantSkus) lines.push(`SKUs: ${variantSkus}`);
    if (variantNames) lines.push(`Variants: ${variantNames}`);
    if (allAttributes) lines.push(`Attributes: ${allAttributes}`);

    // Include any extra metadata fields (except brand/category which are already handled)
    if (metadata) {
      const extraMeta = Object.entries(metadata)
        .filter(([key]) => key !== 'brand' && key !== 'category')
        .map(([key, value]) => `${key}: ${String(value ?? '')}`)
        .join(', ');
      if (extraMeta) lines.push(`Metadata: ${extraMeta}`);
    }

    return lines.join('\n');
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