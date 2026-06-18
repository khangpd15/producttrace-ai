package router

import (
    "github.com/gin-gonic/gin"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
    productHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/handler"
)

type RouterDependency struct {
    BatchHandler   *handler.BatchHandler
    ProductHandler *productHandler.ProductHandler
}

func SetupRouter(deps RouterDependency) *gin.Engine {
    r := gin.Default()

    api := r.Group("api/")

    // Batch routes
    batches := api.Group("batches/")
    {
        batches.GET("/", deps.BatchHandler.GetBatchList)
        batches.GET("/:batch_code", deps.BatchHandler.GetBatchDetail)
        batches.POST("/", deps.BatchHandler.CreateBatch)
    }

    // Product routes
    products := api.Group("products/")
    {
        products.POST("/", deps.ProductHandler.CreateProduct)
        products.GET("/", deps.ProductHandler.GetAllProducts)
        products.GET("/:id", deps.ProductHandler.GetProductByID)
        products.PUT("/:id", deps.ProductHandler.UpdateProduct)
        products.DELETE("/:id", deps.ProductHandler.DeleteProduct)
    }

    return r
}