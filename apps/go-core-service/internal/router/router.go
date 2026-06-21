package router

import (
	"github.com/gin-gonic/gin"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
)

type RouterDependency struct {
	BatchHandler *handler.BatchHandler
}

func SetupRouter(deps RouterDependency) *gin.Engine {
	r := gin.Default()

	api := r.Group("api/")
	SetupBatchRouter(api, deps.BatchHandler)
	return r
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
