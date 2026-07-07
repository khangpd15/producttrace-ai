export const RABBITMQ = {

  URL:
    process.env.RABBITMQ_URL
    ||
    'amqp://admin:admin123@localhost:5672/%2F',


  EXCHANGE:
    'product-trace.events',


  DLX_EXCHANGE:
    'product-trace.dlx',


  EXCHANGE_TYPE:
    'topic',


  QUEUES: {

    USER_REGISTERED:
      'ai.events',

    USER_VERIFIED:
      'notification.user.verified',

    PASSWORD_RESET:
      'notification.password.reset',

    PRODUCT_CREATED:
      'notification.product.created',

  },


  ROUTING_KEYS: {

    USER_REGISTERED:
      'otp.registered',

    USER_VERIFIED:
      'otp.verified',

    PASSWORD_RESET:
      'otp.password_reset_requested',

    PRODUCT_CREATED:
      'product.created',

  }

};