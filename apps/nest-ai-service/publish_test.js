const amqp = require('amqplib');

async function publish() {
  try {
    const connection = await amqp.connect('amqp://admin:admin123@localhost:5672/%2F');
    const channel = await connection.createChannel();
    
    const exchange = 'product-trace.events';
    const routingKey = 'product.created';
    
    const payload = {
      pattern: "product.created",
      data: {
        event_id: `evt-prod-test-${Date.now()}`,
        event_type: "product.created",
        event_version: "1.0",
        timestamp: new Date().toISOString(),
        producer: "go-core-service",
        correlation_id: `corr-prod-test-${Date.now()}`,
        payload: {
          id: "prod-001",
          productId: "prod-001",
          name: "Sản phẩm test Node",
          description: "Mô tả sản phẩm test cho embedding từ Node script",
          slug: "san-pham-test-node",
          status: "ACTIVE",
          createdBy: "admin",
          tags: ["test", "embedding", "node"],
          metadata: { category: "electronics", brand: "Teknix" }
        }
      }
    };
    
    channel.publish(exchange, routingKey, Buffer.from(JSON.stringify(payload)), {
        persistent: true,
        contentType: 'application/json'
    });
    
    console.log('Event published successfully!');
    setTimeout(() => {
      connection.close();
      process.exit(0);
    }, 500);
  } catch (error) {
    console.error('Failed:', error);
    process.exit(1);
  }
}

publish();
