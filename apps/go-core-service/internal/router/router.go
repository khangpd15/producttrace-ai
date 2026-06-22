package router

import (
	"github.com/gin-gonic/gin"
	authHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/handler"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
	batchHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
)

type RouterDependency struct {
	BatchHandler *batchHandler.BatchHandler
	AuthHandler  *authHandler.AuthenHandler
}

func SetupRouter(deps RouterDependency) *gin.Engine {
	r := gin.Default()

	api := r.Group("api")

	SetupBatchRouter(api, deps.BatchHandler)
	SetupAuthRouter(api, deps.AuthHandler)

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

func SetupBatchRouter(api *gin.RouterGroup, batchHandler *handler.BatchHandler) {
	batches := api.Group("/batches")
	{
		batches.GET("", batchHandler.GetBatchList)
		batches.GET("/:batch_code", batchHandler.GetBatchDetail)
		batches.GET("/export-qr/:batch_id", batchHandler.ExportQR)
		batches.POST("", batchHandler.CreateBatch)

	}
}
