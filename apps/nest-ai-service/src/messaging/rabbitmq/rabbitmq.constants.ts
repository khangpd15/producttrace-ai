export const RABBITMQ = {
  URL: process.env.RABBITMQ_URL!,

  QUEUE: 'ai.events',

  DLX: 'product-trace.dlx',

  DLQ_ROUTING_KEY: 'ai.events.failed',

  ROUTING_KEYS: {
    PRODUCT_CREATED: 'product.created',
    USER_REGISTERED: 'user.registered',
    PASSWORD_RESET_REQUESTED: 'auth.password_reset_requested',
  },
};

