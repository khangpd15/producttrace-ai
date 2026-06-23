package router

import (
	"github.com/gin-gonic/gin"

	authHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/handler"
	batchHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
	productHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/handler"
	userHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/handler"
)

type RouterDependency struct {
	BatchHandler   *batchHandler.BatchHandler
	AuthHandler    *authHandler.AuthenHandler
	UserHandler    *userHandler.UserHandler
	ProductHandler *productHandler.ProductHandler
}

func SetupRouter(deps RouterDependency) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")

	SetupAuthRouter(api, deps.AuthHandler)
	SetupUserRouter(api, deps.UserHandler)
	SetupBatchRouter(api, deps.BatchHandler)
	SetupProductRouter(api, deps.ProductHandler)

	return r
}

// AUTH
func SetupAuthRouter(api *gin.RouterGroup, ah *authHandler.AuthenHandler) {
	auth := api.Group("/auth")
	{
		auth.POST("/register", ah.Register)
		auth.POST("/login", ah.Login)
		auth.POST("/verify-otp", ah.VerifyOTP)
		auth.POST("/refresh", ah.RefreshToken)
	}
}

// USER
func SetupUserRouter(api *gin.RouterGroup, uh *userHandler.UserHandler) {
	users := api.Group("/users")
	{
		users.POST("", uh.CreateUser)
		users.PUT("/:id", uh.UpdateUser)
		users.DELETE("/:id", uh.DeleteUser)
		users.GET("", uh.ListUsers)
		users.GET("/:id", uh.GetUserDetail)

		users.GET("/profile", uh.GetProfile)
		users.PUT("/profile/:id", uh.UpdateProfile)
	}
}

// BATCH
func SetupBatchRouter(api *gin.RouterGroup, bh *batchHandler.BatchHandler) {
	batches := api.Group("/batches")
	{
		batches.GET("", bh.GetBatchList)
		batches.GET("/:batch_code", bh.GetBatchDetail)
		batches.GET("/export-qr/:batch_id", bh.ExportQR)
		batches.POST("", bh.CreateBatch)
	}
}

// PRODUCT
func SetupProductRouter(api *gin.RouterGroup, ph *productHandler.ProductHandler) {
	products := api.Group("/products")
	{
		products.POST("", ph.CreateProduct)
		products.GET("", ph.GetAllProducts)
		products.GET("/:id", ph.GetProductByID)
		products.PUT("/:id", ph.UpdateProduct)
		products.DELETE("/:id", ph.DeleteProduct)
	}
}
