package app

import (
	"database/sql"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/qr"
	batchRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/services"
	productItemsRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/router"
	"gorm.io/gorm"
)

type App struct {
	Router *gin.Engine
}

func NewApp(database *sql.DB, dbGORM *gorm.DB) *App {

	productItemsRepo := productItemsRepo.NewProductItemRepository(dbGORM)

	qrGenerator := qr.NewGenerator()
	pdfGenerator := qr.NewPDFGenerator(qrGenerator, os.Getenv("BASE_URL"))

	batchRepo := batchRepo.NewBatchRepository(database)
	batchService := services.NewbatchService(batchRepo, pdfGenerator, productItemsRepo)
	batchHandler := handler.NewBatchHandler(batchService)

	r := router.SetupRouter(router.RouterDependency{
		BatchHandler: batchHandler,
	})

	return &App{
		Router: r,
	}
}
