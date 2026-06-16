package rabbitmq

const (
	// Exchanges
	EventExchange = "product-trace.events"
	DLXExchange   = "product-trace.dlx"

	// Queues
	AIQueue = "ai.events"
	AIDLQ   = "ai.events.dlq"

	// Routing Keys
	ProductCreatedRK = "product.created"
	AIDLQRoutingKey  = "ai.events.failed"
)