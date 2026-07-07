package app

import (
	"database/sql"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	batchHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
	batchRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/qr"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/services"
	productHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/handler"
	productRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/repositories"
	productService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/services"
	productItemsRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/repositories"
	variantRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/router"

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
	auditlog "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/audit_log"
)

type App struct {
	Router *gin.Engine
}

// NewApp bootstraps and configures all components and their DI requirements.
// _ *sql.DB: kept for backward-compatible call sites in cmd/; no longer used
// because all repositories now use GORM exclusively.
func NewApp(database *gorm.DB, redisClient *redis.Client, pub *publisher.Publisher, _ *sql.DB) *App {

	// Audit Log (shared service — inject vào mọi module cần ghi log)
	auditRepo := auditlog.NewAuditLogRepository(database)
	auditService := auditlog.NewAuditLogService(auditRepo)

	productItemsRepo := productItemsRepo.NewProductItemRepository(database)

	// Initialize User Module
	uRepo := userRepo.NewUserRepository(database)
	uService := userService.NewUserService(uRepo, pub)
	uHandler := userHandler.NewUserHandler(uService)
	qrGenerator := qr.NewGenerator()
	pdfGenerator := qr.NewPDFGenerator(qrGenerator, os.Getenv("BASE_URL"))

	bRepo := batchRepo.NewBatchRepository(database)

	// Product module
	pRepo := productRepo.NewProductRepository(database)
	pVariantRepo := variantRepo.NewProductVariantRepository(database)

	batchService := services.NewbatchService(bRepo, pdfGenerator, productItemsRepo, pVariantRepo, pub, auditService)
	batchHandler := batchHandler.NewBatchHandler(batchService)

	// Initialize Auth Module
	redisCache := cache.NewRedisCache(redisClient)
	aService := authService.NewAuthenService(uRepo, redisCache, pub)
	aHandler := authHandler.NewAuthenHandler(aService)

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
