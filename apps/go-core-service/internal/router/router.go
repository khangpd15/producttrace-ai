package router

import (
    "github.com/gin-gonic/gin"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
    productHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/handler"
    authHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/handler"
    userHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/handler"
    batchHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
    locationHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/handler"
)



type RouterDependency struct {
	BatchHandler    *batchHandler.BatchHandler
	AuthHandler     *authHandler.AuthenHandler
	UserHandler     *userHandler.UserHandler
	ProductHandler  *productHandler.ProductHandler
	LocationHandler *locationHandler.LocationHandler
}

func SetupRouter(deps RouterDependency) *gin.Engine {
	r := gin.Default()

	api := r.Group("api")

	SetupBatchRouter(api, deps.BatchHandler)
	SetupAuthRouter(api, deps.AuthHandler)
	SetupUserRouter(api, deps.UserHandler)
	SetupLocationRouter(api, deps.LocationHandler)

	return r
}

func SetupAuthRouter(api *gin.RouterGroup, ah *authHandler.AuthenHandler) {
	auth := api.Group("auth")
	{
		auth.POST("/register", ah.Register)
		auth.POST("/login", ah.Login)
		auth.POST("/verify-otp", ah.VerifyOTP)
		auth.POST("/refresh", ah.RefreshToken)
	}
}

func SetupUserRouter(api *gin.RouterGroup, uh *userHandler.UserHandler) {
	users := api.Group("users")
	{
		users.PUT("/profile/:id", uh.UpdateProfile)
		users.POST("", uh.CreateUser)
		
		users.PUT("/:id", uh.UpdateUser)
		users.DELETE("/:id", uh.DeleteUser)
		users.GET("", uh.ListUsers)
		users.GET("/:id", uh.GetUserDetail)
  }
}

func SetupBatchRouter(api *gin.RouterGroup, batchHandler *handler.BatchHandler) {
	batches := api.Group("/batches")
	{
		batches.GET("", batchHandler.GetBatchList)
		batches.GET("/:batch_code", batchHandler.GetBatchDetail)
		batches.GET("/export-qr/:batch_id", batchHandler.ExportQR)
		batches.POST("", batchHandler.CreateBatch)

	}
}
func SetupProductRouter(api *gin.RouterGroup, productHandler *productHandler.ProductHandler) {
    products := api.Group("/products")
    // Product routes
    {
        products.POST("/", productHandler.CreateProduct)
        products.GET("/", productHandler.GetAllProducts)
        products.GET("/:id", productHandler.GetProductByID)
        products.PUT("/:id", productHandler.UpdateProduct)
        products.DELETE("/:id", productHandler.DeleteProduct)
    }

}

func SetupLocationRouter(api *gin.RouterGroup, locationHandler *locationHandler.LocationHandler) {
	locations := api.Group("/locations")
	{
		locations.POST("/", locationHandler.Create)
		locations.GET("/", locationHandler.GetAll)
		locations.GET("/nearby", locationHandler.FindNearby)
		locations.GET("/:id", locationHandler.GetByID)
		locations.PUT("/:id", locationHandler.Update)
		locations.DELETE("/:id", locationHandler.Delete)
	}
}
