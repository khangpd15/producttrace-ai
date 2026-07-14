import * as fs from 'fs';

const isDocker = fs.existsSync('/.dockerenv');
const defaultRabbitUrl = isDocker
  ? 'amqp://admin:admin123@rabbitmq:5672/%2F'
  : 'amqp://admin:admin123@localhost:5672/%2F';
const envRabbitUrl = process.env.RABBITMQ_URL;
const resolvedRabbitUrl = !isDocker && envRabbitUrl?.includes('rabbitmq')
  ? defaultRabbitUrl
  : envRabbitUrl || defaultRabbitUrl;

export const RABBITMQ = {

  URL: resolvedRabbitUrl,

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

    EMBEDDING: 'embedding_queue',

    EMBEDDING_SYNC: 'embedding_sync_queue',
  },

  DLX: {
    EMBEDDING: 'embedding.dlx',
  },

  DLQ_ROUTING_KEYS: {

    NOTIFICATION: 'ai.events.failed',

    EMBEDDING: 'embedding.failed',
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
    PASSWORD_RESET: 'otp.forgot',        // Go: OTPForgotRK
    USER_VERIFIED: 'otp.verified',      // Go: OTPVerifiedRK
    PRODUCT_CREATED: 'product.created',   // Go: ProductCreatedRK
    OWNERSHIP_OTP: 'otp.ownership',
    TRACE_CREATED: 'trace.created',
    TRACE_EXPORTED: 'trace.exported',
    TRACE_EVENTS: 'trace.*',
    NOTIFICATION_SENT: 'notification.sent', // Warranty update notification

    EMBEDDING_GENERATED: 'embedding.generated',
    EMBEDDING_COMPLETED: 'embedding.completed',
    EMBEDDING_REINDEX_REQUESTED: 'embedding.reindex.requested',
  },

  /**
   * event_type values inside the message body (set by Go publisher).
   * Used by NotificationConsumer to route to the correct mail handler.
   */
  EVENT_TYPES: {
    USER_REGISTERED: 'otp.registered',
    PASSWORD_RESET: 'otp.forgot',
    USER_VERIFIED: 'otp.verified',
    PRODUCT_CREATED: 'product.created',
    OWNERSHIP_OTP: 'otp.ownership',
    NOTIFICATION_SENT: 'notification.sent',
  },
};