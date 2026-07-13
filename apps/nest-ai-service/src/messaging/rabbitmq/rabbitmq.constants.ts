export const RABBITMQ = {

  URL:
    process.env.RABBITMQ_URL ||
    'amqp://admin:admin123@localhost:5672/%2F',

  EXCHANGE: 'product-trace.events',

  DLX_EXCHANGE: 'product-trace.dlx',

  EXCHANGE_TYPE: 'topic',

  /**
   * Single notification queue consumed by NotificationConsumer.
   * Owned and declared by the Go Core Service.
   * NestJS only asserts (passive) and binds routing keys.
   */
  QUEUES: {
    NOTIFICATION: 'ai.events',
  },

  /**
   * Routing keys published by the Go Core Service.
   * These must match Go constants exactly.
   *
   * Go OTP worker publishes "otp.registered" (OTPRegisterUserRK) to ai.events.
   * Go ForgotPassword publishes "otp.forgot" (OTPForgotRK).
   * Go VerifyOTP publishes "otp.verified" (OTPVerifiedRK).
   * Go product service publishes "product.created" (ProductCreatedRK).
   */
  ROUTING_KEYS: {
    USER_REGISTERED: 'otp.registered',   // Go: OTPRegisterUserRK — sent after OTP generation
    PASSWORD_RESET:  'otp.forgot',        // Go: OTPForgotRK
    USER_VERIFIED:   'otp.verified',      // Go: OTPVerifiedRK
    PRODUCT_CREATED: 'product.created',   // Go: ProductCreatedRK
    NOTIFICATION_SENT: 'notification.sent', // Warranty update notification
  },

  /**
   * event_type values inside the message body (set by Go publisher).
   * Used by NotificationConsumer to route to the correct mail handler.
   */
  EVENT_TYPES: {
    USER_REGISTERED: 'otp.registered',
    PASSWORD_RESET:  'otp.forgot',
    USER_VERIFIED:   'otp.verified',
    PRODUCT_CREATED: 'product.created',
    NOTIFICATION_SENT: 'notification.sent',
  },

};