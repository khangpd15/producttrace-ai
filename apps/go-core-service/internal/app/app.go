package app

import (
	"database/sql"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	batchHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/handler"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/qr"
	batchRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/services"
	productHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/handler"
	productRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/repositories"
	productService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/services"
	attributeHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/handler"
	attributeRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/repositories"
	attributeService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/services"
	attributeValueHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/handler"
	attributeValueRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/repositories"
	attributeValueService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/services"
	categoryHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/handler"
	categoryRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/repositories"
	categoryService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/services"
	productItemsHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/handler"
	productItemsRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/repositories"
	productItemsService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/services"
	variantHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/handler"
	variantRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/repositories"
	variantService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/services"
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
	auditlog "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/audit_log"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/cache"
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

	// Product category & attribute modules
	pCategoryRepo := categoryRepo.NewProductCategoryRepository(database)
	cService := categoryService.NewProductCategoryService(pCategoryRepo)  
cHandler := categoryHandler.NewProductCategoryHandler(cService)
	pAttrRepo := attributeRepo.NewAttributeRepository(database)
	pAttrService := attributeService.NewAttributeService(pAttrRepo, pCategoryRepo)
	pAttrHandler := attributeHandler.NewAttributeHandler(pAttrService)

	// Product Attribute Value module (new)
	pAttrValRepo := attributeValueRepo.NewAttributeValueRepository(database)
	pAttrValService := attributeValueService.NewAttributeValueService(database, pAttrValRepo, pVariantRepo, pAttrRepo, pRepo)
	pAttrValHandler := attributeValueHandler.NewAttributeValueHandler(pAttrValService)

	// Product Variant module (new)
	// vService cần thêm attrValRepo để cascade xoá attribute_value khi xoá 1 variant
	vService := variantService.NewProductVariantService(database, pVariantRepo, pAttrValRepo)
	vHandler := variantHandler.NewProductVariantHandler(vService)

	batchService := services.NewbatchService(bRepo, pdfGenerator, productItemsRepo, pVariantRepo, pub, auditService)
	batchHandler := batchHandler.NewBatchHandler(batchService)

	// Initialize Auth Module
	redisCache := cache.NewRedisCache(redisClient)
	aService := authService.NewAuthenService(uRepo, redisCache, pub)
	aHandler := authHandler.NewAuthenHandler(aService)

	// pService cần thêm pAttrValRepo (tạo/xoá attribute value cascade),
	// pAttrRepo (validate attribute_id tồn tại + đúng category khi tạo product
	// kèm variant+attributes), và pCategoryRepo (validate category_id tồn tại
	// khi tạo/cập nhật product)
	pService := productService.NewProductService(database, pRepo, pVariantRepo, pAttrValRepo, pAttrRepo, pCategoryRepo)
	pHandler := productHandler.NewProductHandler(pService)

	piService := productItemsService.NewProductItemService(productItemsRepo, bRepo, pVariantRepo, nil)
	piHandler := productItemsHandler.NewProductItemHandler(piService)

	r := router.SetupRouter(router.RouterDependency{
		BatchHandler:                 batchHandler,
		ProductHandler:               pHandler,
		CategoryHandler:              cHandler,
		UserHandler:                  uHandler,
		AuthHandler:                  aHandler,
		UserRepo:                     uRepo,
		ProductVariantHandler:        vHandler,
		ProductAttributeHandler:      pAttrHandler,
		ProductAttributeValueHandler: pAttrValHandler,
		ProductItemHandler:           piHandler,
	})

	return &App{
		Router: r,
	}
}