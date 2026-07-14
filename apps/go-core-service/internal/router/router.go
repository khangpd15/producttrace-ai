package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/middleware"
	auditHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/audit/handler"
	authHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/handler"
	batchHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
	dashboardHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/dashboard/handler"
	locationHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/handler"
	ownershipHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/handler"
	productHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/handler"
	attributeHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/handler"
	attributeValueHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/handler"
	productCategoryHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/handler"
	productItemHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/handler"
	variantHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/handler"
	publicHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/public/handler"
	traceHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/trace/handler"
	userHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/handler"
	userRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/repository"
	warrantyClaimHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/handler"
)

type RouterDependency struct {
	BatchHandler                 *batchHandler.BatchHandler
	AuthHandler                  *authHandler.AuthenHandler
	AuditHandler                 *auditHandler.AuditHandler
	UserHandler                  *userHandler.UserHandler
	ProductHandler               *productHandler.ProductHandler
	OwnershipHandler             *ownershipHandler.OwnershipHandler
	WarrantyClaimHandler         *warrantyClaimHandler.WarrantyClaimHandler
	UserRepo                     userRepo.UserRepositoryInterface
	LocationHandler              *locationHandler.LocationHandler
	DashboardHandler             *dashboardHandler.DashboardHandler
	TraceHandler                 *traceHandler.TraceHandler
	PublicHandler                *publicHandler.PublicHandler
	RateLimiter                  *middleware.RateLimiter
	ProductCategoryHandler       *productCategoryHandler.ProductCategoryHandler
	ProductVariantHandler        *variantHandler.ProductVariantHandler        // new
	ProductAttributeHandler      *attributeHandler.AttributeHandler           // new
	ProductAttributeValueHandler *attributeValueHandler.AttributeValueHandler // new
	ProductItemHandler           *productItemHandler.ProductItemHandler
}

func SetupRouter(deps RouterDependency) *gin.Engine {
	r := gin.Default()

	// Disable proxy trusting by default to resolve the security warning
	_ = r.SetTrustedProxies(nil)

	// CORS middleware — must be registered before all routes
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Apply global Recovery, RequestID, and Logger middlewares
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())

	api := r.Group("/api")

	SetupAuthRouter(api, deps.AuthHandler)
	SetupUserRouter(api, deps.UserHandler, deps.UserRepo)
	SetupBatchRouter(api, deps.BatchHandler, deps.UserRepo)
	SetupProductRouter(api, deps.ProductHandler, deps.UserRepo)
	SetupLocationRouter(api, deps.LocationHandler, deps.UserRepo)
	SetupProductVariantRouter(api, deps.ProductVariantHandler, deps.UserRepo)               // new
	SetupProductAttributeRouter(api, deps.ProductAttributeHandler, deps.UserRepo)           // new
	SetupProductAttributeValueRouter(api, deps.ProductAttributeValueHandler, deps.UserRepo) // new
	SetupTraceRouter(api, deps.TraceHandler, deps.RateLimiter, deps.UserRepo)
	SetupAuditRouter(api, deps.AuditHandler, deps.UserRepo)
	SetupProductCategoryRouter(api, deps.ProductCategoryHandler, deps.UserRepo)
	return r
}

// AUTH
func SetupAuthRouter(api *gin.RouterGroup, ah *authHandler.AuthenHandler) {
	auth := api.Group("/auth")
	{
		auth.POST("/register", ah.Register)
		auth.POST("/login", ah.Login)
		auth.POST("/verify-otp", ah.VerifyOTP)
		auth.POST("/resend-otp", ah.ResendOTP)
		auth.POST("/refresh", ah.RefreshToken)
		auth.POST("/logout", ah.Logout)
	}
}

// USER
func SetupUserRouter(api *gin.RouterGroup, uh *userHandler.UserHandler, uRepo userRepo.UserRepositoryInterface) {
	users := api.Group("/users")

	// Protected profile routes (requires authentication)
	profileGroup := users.Group("")
	profileGroup.Use(middleware.AuthMiddleware(uRepo))
	{
		profileGroup.GET("/profile", uh.GetProfile)
		profileGroup.PUT("/profile/:id", uh.UpdateProfile)
		profileGroup.GET("/search", uh.SearchUsers)
	}

	// Admin-only management routes (requires ADMIN role)
	adminGroup := users.Group("")
	adminGroup.Use(middleware.AuthMiddleware(uRepo), middleware.RoleMiddleware("ADMIN"))
	{
		adminGroup.POST("", uh.CreateUser)
		adminGroup.PUT("/:id", uh.UpdateUser)
		adminGroup.DELETE("/:id", uh.DeleteUser)
		adminGroup.GET("", uh.ListUsers)
		adminGroup.GET("/:id", uh.GetUserDetail)
		adminGroup.PUT("/:id/lock", uh.LockAccount)
		adminGroup.PUT("/:id/unlock", uh.UnlockAccount)
	}
}

// BATCH
func SetupBatchRouter(api *gin.RouterGroup, bh *batchHandler.BatchHandler, uRepo userRepo.UserRepositoryInterface) {
	batches := api.Group("/batches")

	// Public batch detail route (for anonymous QR scanning)
	batches.GET("/:id", bh.GetBatchDetail)

	// Protected batch routes
	protectedBatches := batches.Group("")
	protectedBatches.Use(middleware.AuthMiddleware(uRepo))
	{
		// ALL AUTHENTICATED USERS can view list and events
		protectedBatches.GET("", bh.GetBatchList)
		// Search batches — UC-P2-BATCH-03: GET /api/v1/batches/search
		// Gin ưu tiên static segment "/search" hơn parameterized "/:id".
		protectedBatches.GET("/search", bh.SearchBatch)
		protectedBatches.GET("/:id/events", bh.GetBatchEvents)

		// MANAGER and WAREHOUSE can export batch
		exportGroup := protectedBatches.Group("")
		exportGroup.Use(middleware.RoleMiddleware("MANAGER", "WAREHOUSE", "ADMIN")) // usually admin has all access
		{
			// Bulk export: POST /batches/export (new) — xuất nhiều batch, toàn bộ ProductItems, all-or-nothing.
			// Đặt trước /:id/export để Gin ưu tiên static segment.
			exportGroup.POST("/export", bh.ExportBatches)
			// Legacy single-batch export: POST /batches/:id/export — giữ lại để backward compat.
			exportGroup.POST("/:id/export", bh.ExportBatch)
		}


		// ADMIN and MANUFACTURER roles can export QR PDF, and create/update/delete batches
		staffGroup := protectedBatches.Group("")
		staffGroup.Use(middleware.RoleMiddleware("ADMIN", "MANUFACTURER"))
		{
			staffGroup.GET("/export-qr/:id", bh.ExportQR)
			staffGroup.POST("", bh.CreateBatch)
			staffGroup.PATCH("/:id/status", bh.UpdateBatchStatus)
			staffGroup.DELETE("/:id", bh.DeleteBatch)
		}
	}
}

// PRODUCT
func SetupProductRouter(api *gin.RouterGroup, ph *productHandler.ProductHandler, uRepo userRepo.UserRepositoryInterface) {
	products := api.Group("/products")
	{
		// Public endpoints to browse products
		products.GET("", ph.GetAllProducts)
		products.GET("/:id", ph.GetProductByID)

		// Protected product management routes
		protectedProducts := products.Group("")
		protectedProducts.Use(middleware.AuthMiddleware(uRepo))
		{
			// ADMIN or MANUFACTURER can create or update products
			staffGroup := protectedProducts.Group("")
			staffGroup.Use(middleware.RoleMiddleware("ADMIN", "MANUFACTURER"))
			{
				staffGroup.POST("", ph.CreateProduct)
				staffGroup.PUT("/:id", ph.UpdateProduct)
			}

			// Only ADMIN can delete products
			adminGroup := protectedProducts.Group("")
			adminGroup.Use(middleware.RoleMiddleware("ADMIN"))
			{
				adminGroup.DELETE("/:id", ph.DeleteProduct)
			}
		}
	}
}

// OWNERSHIP
func SetupOwnershipRouter(api *gin.RouterGroup, oh *ownershipHandler.OwnershipHandler, uRepo userRepo.UserRepositoryInterface) {
	ownerships := api.Group("/ownership")
	ownerships.Use(middleware.AuthMiddleware(uRepo))

	// Customer routes: chỉ cần QR code, thông tin lấy từ profile JWT
	customerGroup := ownerships.Group("")
	customerGroup.Use(middleware.RoleMiddleware("CUSTOMER"))
	{
		customerGroup.POST("/request-otp", oh.CustomerRequestOTP)
		customerGroup.POST("/register", oh.CustomerVerifyAndRegister)
	}

	// Admin routes: Admin điền đầy đủ thông tin thay cho khách hàng
	adminGroup := ownerships.Group("/admin")
	adminGroup.Use(middleware.RoleMiddleware("ADMIN"))
	{
		adminGroup.POST("/request-otp", oh.AdminRequestOTP)
		adminGroup.POST("/register", oh.AdminVerifyAndRegister)
	}

	// Detail route: Tất cả user đã auth đều có thể xem thông tin sở hữu
	ownerships.GET("/detail/:product_item_id", oh.GetOwnershipDetail) // updated to avoid conflict with /:id

	// CRUD Extensions
	ownerships.PUT("/:id/transfer", oh.TransferOwnership)
	ownerships.DELETE("/:id", oh.DeleteOwnership)
	ownerships.GET("", oh.SearchOwnerships)
}

// WARRANTY CLAIM
func SetupWarrantyClaimRouter(api *gin.RouterGroup, wch *warrantyClaimHandler.WarrantyClaimHandler, uRepo userRepo.UserRepositoryInterface) {
	warrantyClaims := api.Group("/warranty-claims")
	warrantyClaims.Use(middleware.AuthMiddleware(uRepo))
	{
		warrantyClaims.POST("", wch.CreateWarrantyClaim)
	}
}

func SetupLocationRouter(api *gin.RouterGroup, locationHandler *locationHandler.LocationHandler, uRepo userRepo.UserRepositoryInterface) {
	locations := api.Group("/locations")
	{
		// Public endpoints to browse locations
		locations.GET("", locationHandler.GetAll)
		locations.GET("/:id", locationHandler.GetByID)

		// Admin-only management routes (requires ADMIN role)
		adminGroup := locations.Group("")
		adminGroup.Use(middleware.AuthMiddleware(uRepo), middleware.RoleMiddleware("ADMIN"))
		{
			adminGroup.POST("", locationHandler.Create)
			adminGroup.PUT("/:id", locationHandler.Update)
			adminGroup.DELETE("/:id", locationHandler.Delete)
		}
	}
}

func SetupDashboardRouter(api *gin.RouterGroup, dh *dashboardHandler.DashboardHandler, uRepo userRepo.UserRepositoryInterface) {
	dashboard := api.Group("/dashboard")
	dashboard.Use(middleware.AuthMiddleware(uRepo), middleware.RoleMiddleware("ADMIN", "STAFF"))
	{
		dashboard.GET("/stats", dh.GetStats)
		dashboard.GET("/activities", dh.GetActivities)
		dashboard.GET("/alerts", dh.GetAlerts)
		dashboard.GET("/charts/production-sales", dh.GetProductionSalesChart)
	}
}

// PRODUCT VARIANT
func SetupProductVariantRouter(api *gin.RouterGroup, vh *variantHandler.ProductVariantHandler, uRepo userRepo.UserRepositoryInterface) {
	variants := api.Group("/variants")
	{
		// Public endpoints to browse variants
		variants.GET("/:id", vh.GetVariantByID)
		variants.GET("/product/:product_id", vh.GetVariantsByProductID)

		// Protected variant management routes
		protectedVariants := variants.Group("")
		protectedVariants.Use(middleware.AuthMiddleware(uRepo))
		{
			// ADMIN or MANUFACTURER can update variants
			staffGroup := protectedVariants.Group("")
			staffGroup.Use(middleware.RoleMiddleware("ADMIN", "MANUFACTURER"))
			{
				staffGroup.PUT("/:id", vh.UpdateVariant)
			}

			// Only ADMIN can delete variants
			adminGroup := protectedVariants.Group("")
			adminGroup.Use(middleware.RoleMiddleware("ADMIN"))
			{
				adminGroup.DELETE("/:id", vh.DeleteVariant)
			}
		}
	}
}

// PRODUCT ATTRIBUTE
func SetupProductAttributeRouter(api *gin.RouterGroup, ah *attributeHandler.AttributeHandler, uRepo userRepo.UserRepositoryInterface) {
	attributes := api.Group("/attributes")
	{
		// Public endpoints to browse attributes
		attributes.GET("", ah.ListAttributes)
		attributes.GET("/:id", ah.GetAttributeByID)

		// Protected attribute management routes
		protectedAttributes := attributes.Group("")
		protectedAttributes.Use(middleware.AuthMiddleware(uRepo))
		{
			// ADMIN or MANUFACTURER can create or update attributes
			staffGroup := protectedAttributes.Group("")
			staffGroup.Use(middleware.RoleMiddleware("ADMIN", "MANUFACTURER"))
			{
				staffGroup.POST("", ah.CreateAttribute)
				staffGroup.PUT("/:id", ah.UpdateAttribute)
			}

			// Only ADMIN can delete attributes
			adminGroup := protectedAttributes.Group("")
			adminGroup.Use(middleware.RoleMiddleware("ADMIN"))
			{
				adminGroup.DELETE("/:id", ah.DeleteAttribute)
			}
		}
	}
}

// PRODUCT ATTRIBUTE VALUE
func SetupProductAttributeValueRouter(api *gin.RouterGroup, ah *attributeValueHandler.AttributeValueHandler, uRepo userRepo.UserRepositoryInterface) {
	values := api.Group("/attribute-values")
	{
		values.GET("", ah.ListAllAttributeValues)
		values.GET("/:id", ah.GetAttributeValueByID)

		protectedValues := values.Group("")
		protectedValues.Use(middleware.AuthMiddleware(uRepo))
		{
			staffGroup := protectedValues.Group("")
			staffGroup.Use(middleware.RoleMiddleware("ADMIN", "MANUFACTURER"))
			{
				staffGroup.PUT("/:id", ah.UpdateAttributeValue)
			}

			adminGroup := protectedValues.Group("")
			adminGroup.Use(middleware.RoleMiddleware("ADMIN"))
			{
				adminGroup.DELETE("/:id", ah.DeleteAttributeValue)
			}
		}
	}

	variants := api.Group("/variants")
	{
		variants.GET("/:id/attributes", ah.GetAttributeValuesByVariantID)

		protectedVariants := variants.Group("")
		protectedVariants.Use(middleware.AuthMiddleware(uRepo))
		{
			staffGroup := protectedVariants.Group("")
			staffGroup.Use(middleware.RoleMiddleware("ADMIN", "MANUFACTURER"))
			{
				staffGroup.POST("/:id/attributes", ah.AssignAttributes)
			}
		}
	}
}

// TRACE
func SetupTraceRouter(api *gin.RouterGroup, th *traceHandler.TraceHandler, rl *middleware.RateLimiter, uRepo userRepo.UserRepositoryInterface) {

	legacy := api.Group("/trace")

	// Public search with rate limiting
	legacy.GET("/search", rl.Limit(30, time.Minute), th.Search)

	// Protected export endpoints
	protected := legacy.Group("")
	protected.Use(middleware.AuthMiddleware(uRepo))
	{
		protected.POST("/export/pdf", th.ExportPDF)
		protected.POST("/export/excel", th.ExportExcel)
	}

}

// PRODUCT CATEGORY
func SetupProductCategoryRouter(api *gin.RouterGroup, ch *productCategoryHandler.ProductCategoryHandler, uRepo userRepo.UserRepositoryInterface) {
	categories := api.Group("/categories")
	{
		// Public browse routes
		categories.GET("", ch.GetAllCategories)
		categories.GET("/:id", ch.GetCategoryByID)

		// Protected management routes (requires ADMIN or STAFF role)
		protected := categories.Group("")
		protected.Use(middleware.AuthMiddleware(uRepo), middleware.RoleMiddleware("ADMIN", "STAFF"))
		{
			protected.POST("", ch.CreateCategory)
			protected.PUT("/:id", ch.UpdateCategory)
			protected.DELETE("/:id", ch.DeleteCategory)
		}
	}
}

// AUDIT
func SetupAuditRouter(api *gin.RouterGroup, ah *auditHandler.AuditHandler, uRepo userRepo.UserRepositoryInterface) {
	audits := api.Group("/audits")
	audits.Use(middleware.AuthMiddleware(uRepo), middleware.RoleMiddleware("ADMIN"))
	{
		audits.GET("", ah.GetLogs)
	}
}
