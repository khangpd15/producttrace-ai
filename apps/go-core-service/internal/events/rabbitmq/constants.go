package rabbitmq

const (
	// Exchanges
	DefaultExchange = "product-trace.events"
	DLXExchange     = "product-trace.dlx"

	// Alias for compatibility/readability
	EventExchange = DefaultExchange

	// User routing keys
	UserRegisteredRK     = "user.registered"
	UserPasswordForgotRK = "user.password_forgot"
	UserLoggedInRK       = "user.logged_in"

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
	TraceCreatedRK  = "trace.created"
	TraceExportedRK = "trace.exported"

	// Notification routing keys
	NotificationCreatedRK = "notification.created"
	OTPRegisterUserRK     = "otp.registered"
	OTPVerifiedRK         = "otp.verified"
	OTPForgotRK           = "otp.forgot"
	OTPOwnership          = "otp.ownership"
)
