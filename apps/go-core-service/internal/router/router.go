package router

import (
	"github.com/gin-gonic/gin"
	authHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/handler"
	batchHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
	userHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/handler"
)

type RouterDependency struct {
	BatchHandler *batchHandler.BatchHandler
	AuthHandler  *authHandler.AuthenHandler
	UserHandler  *userHandler.UserHandler
}

func SetupRouter(deps RouterDependency) *gin.Engine {
	r := gin.Default()

	api := r.Group("api")
	
	SetupBatchRouter(api, deps.BatchHandler)
	SetupAuthRouter(api, deps.AuthHandler)
	SetupUserRouter(api, deps.UserHandler)

	return r
}

func SetupBatchRouter(api *gin.RouterGroup, bh *batchHandler.BatchHandler) {
	batches := api.Group("batches")
	{
		batches.GET("", bh.GetBatchList)
		batches.GET("/:batch_code", bh.GetBatchDetail)
		batches.POST("", bh.CreateBatch)
	}
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
