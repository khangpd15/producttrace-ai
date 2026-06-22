package app

import (
    "database/sql"
    "log"

    "github.com/gin-gonic/gin"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    batchHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
    batchRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/repositories"
    batchService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/services"
    productHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/handler"
    productRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/repositories"
    productService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/services"
    variantRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/repositories"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/router"
)

type App struct {
    Router *gin.Engine
}

func NewApp(database *sql.DB) *App {
    // Batch module
    batchR := batchRepo.NewBatchRepository(database)
    batchS := batchService.NewbatchService(batchR)
    batchH := batchHandler.NewBatchHandler(batchS)

    // Wrap *sql.DB thành GORM
    gormDB, err := gorm.Open(postgres.New(postgres.Config{
        Conn: database,
    }), &gorm.Config{})
    if err != nil {
        log.Fatalf("failed to init gorm: %v", err)
    }

    // Product module
    pRepo := productRepo.NewProductRepository(gormDB)
    pVariantRepo := variantRepo.NewProductVariantRepository(gormDB)
    pService := productService.NewProductService(gormDB, pRepo, pVariantRepo)
    pHandler := productHandler.NewProductHandler(pService)

    r := router.SetupRouter(router.RouterDependency{
        BatchHandler:   batchH,
        ProductHandler: pHandler,
    })

    return &App{
        Router: r,
    }
}