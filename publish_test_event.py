import json
import pika

payload = {
  "pattern": "product.created",
  "data": {
    "event_id": "evt-prod-test-001",
    "event_type": "product.created",
    "event_version": "1.0",
    "timestamp": "2026-07-12T10:00:00Z",
    "producer": "go-core-service",
    "correlation_id": "corr-prod-test-001",
    "payload": {
      "id": "prod-001",
      "productId": "prod-001",
      "name": "Sản phẩm test",
      "description": "Mô tả sản phẩm test cho embedding",
      "slug": "san-pham-test",
      "status": "ACTIVE",
      "createdBy": "admin",
      "tags": ["test", "embedding"],
      "metadata": {"category": "electronics", "brand": "Teknix"}
    }
  }
}

connection = pika.BlockingConnection(
    pika.ConnectionParameters(host='localhost', port=5672, credentials=pika.PlainCredentials('admin', 'admin123'))
)
channel = connection.channel()
channel.exchange_declare(exchange='product-trace.events', exchange_type='topic', durable=True)
channel.basic_publish(exchange='product-trace.events', routing_key='product.created', body=json.dumps(payload))
connection.close()
print('published')
