export const RABBITMQ = {
  URL: process.env.RABBITMQ_URL!,

  QUEUE: 'ai.events',

  DLX: 'product-trace.dlx',

  DLQ_ROUTING_KEY: 'ai.events.failed',

  ROUTING_KEYS: {
    PRODUCT_CREATED: 'product.created',
  },
};

