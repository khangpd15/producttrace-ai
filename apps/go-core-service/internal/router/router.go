package router

import (
	"github.com/gin-gonic/gin"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/middleware"
	authHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/handler"
	batchHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
	locationHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/handler"
	productHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/handler"
	userHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/handler"
	userRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/repository"
)

type RouterDependency struct {
	BatchHandler    *batchHandler.BatchHandler
	AuthHandler     *authHandler.AuthenHandler
	UserHandler     *userHandler.UserHandler
	ProductHandler  *productHandler.ProductHandler
	UserRepo        userRepo.UserRepositoryInterface
	LocationHandler *locationHandler.LocationHandler
}

func SetupRouter(deps RouterDependency) *gin.Engine {
	r := gin.Default()

	// Disable proxy trusting by default to resolve the security warning
	_ = r.SetTrustedProxies(nil)

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
	batches.GET("/:batch_code", bh.GetBatchDetail)

	// Protected batch routes
	protectedBatches := batches.Group("")
	protectedBatches.Use(middleware.AuthMiddleware(uRepo))
	{
		// ADMIN and MANUFACTURER roles can view list, export QR PDF, and create batches
		staffGroup := protectedBatches.Group("")
		staffGroup.Use(middleware.RoleMiddleware("ADMIN", "MANUFACTURER"))
		{
			staffGroup.GET("", bh.GetBatchList)
			staffGroup.GET("/export-qr/:batch_id", bh.ExportQR)
			staffGroup.POST("", bh.CreateBatch)
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
		locations.GET("/", locationHandler.GetAll)
		locations.GET("/:id", locationHandler.GetByID)

		// Admin-only management routes (requires ADMIN role)
		adminGroup := locations.Group("")
		adminGroup.Use(middleware.AuthMiddleware(uRepo), middleware.RoleMiddleware("ADMIN"))
		{
			adminGroup.POST("/", locationHandler.Create)
			adminGroup.PUT("/:id", locationHandler.Update)
			adminGroup.DELETE("/:id", locationHandler.Delete)
		}
	}
}
