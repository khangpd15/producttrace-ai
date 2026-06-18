package app

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/services"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/router"
)

type App struct {
	Router *gin.Engine
}

func NewApp(database *sql.DB) *App {
	batchRepo := repositories.NewBatchRepository(database)
	batchService := services.NewbatchService(batchRepo)
	batchHandler := handler.NewBatchHandler(batchService)

	r := router.SetupRouter(router.RouterDependency{
		BatchHandler: batchHandler,
	})

	return &App{
		Router: r,
	}
}
