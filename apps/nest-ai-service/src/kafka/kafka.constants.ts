export const KAFKA = {
  BROKERS: (process.env.KAFKA_BROKER ?? 'kafka:9092')
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean),

  GROUP_ID:
    process.env.KAFKA_GROUP_ID ??
    'nest-ai-service-group',

  TOPICS: {
    PRODUCT_EVENTS: 'product-events',
    TRACE_EVENTS: 'trace-events',

    EMBEDDING_GENERATED:
      'embedding.generated',

    EMBEDDING_COMPLETED:
      'embedding.completed',

    EMBEDDING_REINDEX_REQUESTED:
      'embedding.reindex.requested',

    EMBEDDING_REINDEXED:
      'embedding.reindexed',
  }
};
console.log('KAFKA_BROKER =', process.env.KAFKA_BROKER);
console.log('KAFKA.BROKERS =', KAFKA.BROKERS);