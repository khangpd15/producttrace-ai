package app

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"

	batchHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
	batchRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/repositories"
	productHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/handler"
	productRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/repositories"
	productService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/services"
	variantRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/router"

	"github.com/redis/go-redis/v9"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/publisher"

	// Auth Module
	authHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/handler"
	authService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/service"

	// User Module
	userHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/handler"
	userRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/repository"
	userService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/service"

	// Cache pkg
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/cache"

	"os"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/qr"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/services"
	productItemsRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/repositories"
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

	productItemsRepo := productItemsRepo.NewProductItemRepository(database, databaseSQL)

	// Initialize User Module
	uRepo := userRepo.NewUserRepository(database)
	uService := userService.NewUserService(uRepo, pub)
	uHandler := userHandler.NewUserHandler(uService)
	qrGenerator := qr.NewGenerator()
	pdfGenerator := qr.NewPDFGenerator(qrGenerator, os.Getenv("BASE_URL"))

	batchRepo := batchRepo.NewBatchRepository(databaseSQL)
	batchService := services.NewbatchService(batchRepo, pdfGenerator, productItemsRepo)
	batchHandler := batchHandler.NewBatchHandler(batchService)

	// Initialize Auth Module
	redisCache := cache.NewRedisCache(redisClient)
	aService := authService.NewAuthenService(uRepo, redisCache, pub)
	aHandler := authHandler.NewAuthenHandler(aService)

	// Product module
	pRepo := productRepo.NewProductRepository(database)
	pVariantRepo := variantRepo.NewProductVariantRepository(database)
	pService := productService.NewProductService(database, pRepo, pVariantRepo)
	pHandler := productHandler.NewProductHandler(pService)

	r := router.SetupRouter(router.RouterDependency{
		BatchHandler:   batchHandler,
		ProductHandler: pHandler,
		UserHandler:    uHandler,
		AuthHandler:    aHandler,
		UserRepo:       uRepo,
	})

	return &App{
		Router: r,
	}
}
