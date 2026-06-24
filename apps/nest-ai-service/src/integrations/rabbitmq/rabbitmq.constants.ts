export const RABBITMQ = {
  URL:
    process.env.RABBITMQ_URL ||
    'amqp://guest:guest@localhost:5672',

  QUEUES: {
    AI_EVENTS: 'ai.events',
    EMBEDDING: 'embedding_queue',
  },

  DLX: {
    AI_EVENTS: 'product-trace.dlx',
    EMBEDDING: 'embedding_dlx',
  },

  DLQ_ROUTING_KEYS: {
    AI_EVENTS: 'ai.events.failed',
    EMBEDDING: 'embedding_dlq',
  },

  ROUTING_KEYS: {
    PRODUCT_CREATED: 'product.created',
  },
};