package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/middleware"
	authHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/handler"
	batchHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
	locationHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/handler"
	productHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/handler"
	attributeHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/handler"
	attributeValueHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/handler"
	productItemHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/handler"
	variantHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/handler"
	traceHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/trace/handler"
	userHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/handler"
	userRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/repository"
	dashboardHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/dashboard/handler"
	publicHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/public/handler"
)

type RouterDependency struct {
	BatchHandler                 *batchHandler.BatchHandler
	AuthHandler                  *authHandler.AuthenHandler
	UserHandler                  *userHandler.UserHandler
	ProductHandler               *productHandler.ProductHandler
	UserRepo                     userRepo.UserRepositoryInterface
	LocationHandler              *locationHandler.LocationHandler
	DashboardHandler             *dashboardHandler.DashboardHandler
	ProductVariantHandler        *variantHandler.ProductVariantHandler        // new
	ProductAttributeHandler      *attributeHandler.AttributeHandler           // new
	ProductAttributeValueHandler *attributeValueHandler.AttributeValueHandler // new
	ProductItemHandler           *productItemHandler.ProductItemHandler
	TraceHandler                 *traceHandler.TraceHandler
	PublicHandler                *publicHandler.PublicHandler
	RateLimiter                  *middleware.RateLimiter
}

func SetupRouter(deps RouterDependency) *gin.Engine {
	r := gin.Default()

	// Disable proxy trusting by default to resolve the security warning
	_ = r.SetTrustedProxies(nil)

	// CORS is handled centrally by Kong Gateway (see infra/kong/kong.yml).
	// // We disable Go's CORS middleware here to prevent duplicate CORS headers and 403 Forbidden on production origins.
	// r.Use(middleware.CORSMiddleware())
	// r.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
	// 	AllowCredentials: true,
	// 	MaxAge:           12 * time.Hour,
	// }))

	// Apply global Recovery, RequestID, and Logger middlewares
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())

	// Static serving of exported certificates/files
	r.Static("/storage", "./storage")

	api := r.Group("/api")

	SetupAuthRouter(api, deps.AuthHandler)
	SetupUserRouter(api, deps.UserHandler, deps.UserRepo)
	SetupBatchRouter(api, deps.BatchHandler, deps.UserRepo)
	SetupProductRouter(api, deps.ProductHandler, deps.UserRepo)
	// SetupProductItemRouter(api, deps.ProductItemHandler, deps.UserRepo)
	SetupLocationRouter(api, deps.LocationHandler, deps.UserRepo)
	SetupDashboardRouter(api, deps.DashboardHandler, deps.UserRepo)
	SetupProductVariantRouter(api, deps.ProductVariantHandler, deps.UserRepo)               // new
	SetupProductAttributeRouter(api, deps.ProductAttributeHandler, deps.UserRepo)           // new
	SetupProductAttributeValueRouter(api, deps.ProductAttributeValueHandler, deps.UserRepo) // new
	SetupTraceRouter(api, deps.TraceHandler, deps.RateLimiter, deps.UserRepo)
	SetupPublicRouter(api, deps.PublicHandler)
	return r
}

// PUBLIC
func SetupPublicRouter(api *gin.RouterGroup, ph *publicHandler.PublicHandler) {
	public := api.Group("/public")
	{
		public.GET("/verify", ph.VerifyQR)
	}
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
		auth.POST("/forgot-password", ah.ForgotPassword)
		auth.POST("/reset-password", ah.ResetPassword)
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
		// UC-P2-BATCH-05: Xem sản phẩm trong lô — Admin, Staff, Dealer
		protectedBatches.GET("/:id/products", bh.GetBatchProducts)

		// UC-P2-BATCH-06: Xem lịch sử thay đổi lô — chỉ Admin và Staff (không có Dealer/Customer)
		historyGroup := protectedBatches.Group("")
		historyGroup.Use(middleware.RoleMiddleware("ADMIN", "STAFF"))
		{
			historyGroup.GET("/:id/history", bh.GetBatchHistory)
		}

		// MANAGER and WAREHOUSE can export batch
		exportGroup := protectedBatches.Group("")
		exportGroup.Use(middleware.RoleMiddleware("MANAGER", "WAREHOUSE", "ADMIN")) // usually admin has all access
		{
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

// PRODUCT ITEM
func SetupProductItemRouter(api *gin.RouterGroup, ph *productItemHandler.ProductItemHandler, uRepo userRepo.UserRepositoryInterface) {
	items := api.Group("/product-items")
	{
		// Public: dùng để scan QR
		items.GET("", ph.GetProductItemList)
		items.GET("/:item_code", ph.GetProductItemDetail)
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
