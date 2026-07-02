export const RABBITMQ = {
  // Fallback URL if not provided in environment variables
  URL: process.env.RABBITMQ_URL || 'amqp://admin:admin123@localhost:5672/%2F',

  // Exchange configuration
  EXCHANGE: 'producttrace.events',
  EXCHANGE_TYPE: 'topic',

  // Queues configurations
  QUEUES: {
    USER_REGISTERED: 'notification.user.registered',
    USER_VERIFIED: 'notification.user.verified',
    PASSWORD_RESET: 'notification.password.reset',
    PRODUCT_CREATED: 'notification.product.created',
  },

  // Routing Keys configurations
  ROUTING_KEYS: {
    USER_REGISTERED: 'user.registered',
    USER_VERIFIED: 'user.verified',
    PASSWORD_RESET: 'password.reset',
    PRODUCT_CREATED: 'product.created',
  },
};
