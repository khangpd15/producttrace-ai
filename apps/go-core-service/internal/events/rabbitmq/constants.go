package rabbitmq

const (
	// Exchanges
	DefaultExchange = "product-trace.events"
	DLXExchange     = "product-trace.dlx"

	// Alias for compatibility/readability
	EventExchange = DefaultExchange

	// User routing keys
	UserRegisteredRK    = "user.registered"
	UserVerifiedRK      = "user.verified"
	UserPasswordResetRK = "user.password_reset_requested"
	UserLoggedInRK      = "user.logged_in"

	// Product routing keys
	ProductCreatedRK = "product.created"
	ProductUpdatedRK = "product.updated"
	ProductDeletedRK = "product.deleted"

	// Batch routing keys
	BatchCreatedRK = "batch.created"
	BatchUpdatedRK = "batch.updated"
	BatchDeletedRK = "batch.deleted"

	// Owner routing keys
	OwnerCreatedRK = "owner.created"
	OwnerUpdatedRK = "owner.updated"
	OwnerDeletedRK = "owner.deleted"

	// Warranty routing keys
	WarrantyCreatedRK = "warranty.created"

	// Trace routing keys
	TraceCreatedRK = "trace.created"

	// Notification routing keys
	NotificationCreatedRK = "notification.created"
	OTPGeneratedRK        = "otp.generated"
)
