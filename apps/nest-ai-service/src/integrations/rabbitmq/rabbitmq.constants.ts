export const RABBITMQ = {
  URL:
    process.env.RABBITMQ_URL ||
    'amqp://guest:guest@localhost:5672',

  QUEUES: {
    // Queue của Email (không đụng)
    AI_EVENTS: 'ai.events',

    // Queue riêng cho Embedding
    EMBEDDING: 'embedding_queue',
  },

  EXCHANGE: 'product-trace.events',
  EXCHANGE_TYPE: 'topic',

  DLX: {
    AI_EVENTS: 'product-trace.dlx',
    EMBEDDING: 'embedding.dlx',
  },

  DLQ_ROUTING_KEYS: {
    AI_EVENTS: 'ai.events.failed',
    EMBEDDING: 'embedding.failed',
  },

  ROUTING_KEYS: {
    // Core events
    PRODUCT_CREATED: 'product.created',
    TRACE_EVENTS: 'trace.events',

    USER_REGISTERED: 'user.registered',
    PASSWORD_RESET_REQUESTED:
      'auth.password_reset_requested',

    // Embedding pipeline
    EMBEDDING_GENERATED:
      'embedding.generated',

    EMBEDDING_COMPLETED:
      'embedding.completed',

    EMBEDDING_REINDEX_REQUESTED:
      'embedding.reindex.requested',
  },
};