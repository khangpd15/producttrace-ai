package app

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/publisher"

	// Auth Module
	authHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/handler"
	authService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/service"

	// User Module
	userRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/repository"

	// Cache pkg
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/cache"

	// Batch Module
	batchHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
	batchRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/repositories"

	"database/sql"
	"os"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/qr"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/services"
	productItemsRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/router"
	"gorm.io/gorm"
)

type App struct {
	Router *gin.Engine
}

// NewApp bootstraps and configures all components and their DI requirements.
func NewApp(database *gorm.DB, redisClient *redis.Client, pub *publisher.Publisher, databaseSQL *sql.DB) *App {
	// Extract underlying sql.DB for the raw SQL queries in the Batch repository
	_, err := database.DB()
	if err != nil {
		log.Fatalf("failed to retrieve sql.DB from GORM client: %v", err)
	}

	productItemsRepo := productItemsRepo.NewProductItemRepository(database)

	qrGenerator := qr.NewGenerator()
	pdfGenerator := qr.NewPDFGenerator(qrGenerator, os.Getenv("BASE_URL"))

	batchRepo := batchRepo.NewBatchRepository(databaseSQL)
	batchService := services.NewbatchService(batchRepo, pdfGenerator, productItemsRepo)
	batchHandler := batchHandler.NewBatchHandler(batchService)

	uRepo := userRepo.NewUserRepository(database)
	// 3. Initialize Auth Module
	redisCache := cache.NewRedisCache(redisClient)
	aService := authService.NewAuthenService(uRepo, redisCache, pub)
	aHandler := authHandler.NewAuthenHandler(aService)

	// 4. Setup Router
	r := router.SetupRouter(router.RouterDependency{
		BatchHandler: batchHandler,
		AuthHandler:  aHandler,
	})

	return &App{
		Router: r,
	}
}
