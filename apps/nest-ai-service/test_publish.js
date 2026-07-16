const amqp = require('amqplib');

async function run() {
  const conn = await amqp.connect('amqp://admin:admin123@localhost:5672');
  const ch = await conn.createChannel();
  
  const exchange = 'product-trace.events';
  const routingKey = 'embedding.reindex.requested';
  
  const payload = {
    "event_id": "333e4567-e89b-12d3-a456-426614174003",
    "event_type": "embedding.reindex.requested",
    "payload": {}
  };

  ch.publish(exchange, routingKey, Buffer.from(JSON.stringify(payload)));
  console.log('Message published');
  
  setTimeout(() => {
    conn.close();
  }, 500);
}

run().catch(console.error);
