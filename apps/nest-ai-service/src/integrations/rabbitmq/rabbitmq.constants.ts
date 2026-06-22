export const RABBITMQ = {
  URL: process.env.RABBITMQ_URL || 'amqp://guest:guest@localhost:5672',
  QUEUE: 'embedding_queue',
  DLX: 'embedding_dlx',
  DLQ_ROUTING_KEY: 'embedding_dlq',
};