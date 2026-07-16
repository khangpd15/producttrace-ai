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
	auditHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/audit/handler"
	authHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/handler"
	authService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/service"

	// User Module
	userHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/handler"
	userRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/repository"
	userService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/service"

	// Cache pkg
	auditlog "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/audit_log"

	// Location Module

	locationHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/handler"
	locationRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/repository"
	locationService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/service"

	//
	ownershipAdapters "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/adapters"
	ownershipEntity "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/entity"
	ownershipHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/handler"
	ownershipRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/repository"
	ownershipService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/service"

	warrantyEntity "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/entity"
	warrantyHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/handler"
	warrantyRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/repository"
	warrantyService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/service"

	warrantyClaimEntity "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/entity"
	warrantyClaimHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/handler"
	warrantyClaimRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/repository"
	warrantyClaimService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/service"

	// Dashboard Module
	dashboardHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/dashboard/handler"
	dashboardRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/dashboard/repositories"
	dashboardService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/dashboard/services"

	// Trace
	traceHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/trace/handler"
	traceRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/trace/repositories"
	traceService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/trace/services"

	// Middleware
	middleware "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/middleware"
	// Public
	publicHandler "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/public/handler"
	publicService "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/public/service"

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
	auditHandler := auditHandler.NewAuditHandler(auditService)

	productItemsRepo := productItemsRepo.NewProductItemRepository(database)

	// Initialize User Module
	uRepo := userRepo.NewUserRepository(database)
	pRepo := productRepo.NewProductRepository(database)
	uService := userService.NewUserService(uRepo, pRepo, pub, auditService)
	uHandler := userHandler.NewUserHandler(uService)
	qrGenerator := qr.NewGenerator()
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = string("https://frontend-producttrace-ai.vercel.app")
	}
	pdfGenerator := qr.NewPDFGenerator(qrGenerator, frontendURL)

	bRepo := batchRepo.NewBatchRepository(database)

	// Product module
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

	// Initialize Dashboard Module
	dbRepo := dashboardRepo.NewDashboardRepository(database)
	dbService := dashboardService.NewDashboardService(dbRepo)
	dbHandler := dashboardHandler.NewDashboardHandler(dbService)

	piService := productItemsService.NewProductItemService(productItemsRepo, bRepo, pVariantRepo, nil)
	piHandler := productItemsHandler.NewProductItemHandler(piService)

	// Location module
	lRepo := locationRepo.NewLocationRepository(database)
	lService := locationService.NewLocationService(lRepo)
	lHandler := locationHandler.NewLocationHandler(lService)
	// Initialize Trace Module
	tRepo := traceRepo.NewTraceRepository(database)
	tService := traceService.NewTraceService(tRepo, redisClient, pub, auditService, os.Getenv("BASE_URL"))
	tHandler := traceHandler.NewTraceHandler(tService)
	rateLimiter := middleware.NewRateLimiter(redisClient)

	// Ownership Module
	oRepo := ownershipRepo.NewOwnershipRepository(database)
	oProductAdapter := ownershipAdapters.NewRealProductAdapter(database, productItemsRepo, pRepo)
	oEmailAdapter := ownershipAdapters.NewRealEmailAdapter()
	oUserAdapter := ownershipAdapters.NewRealUserAdapter(uRepo)
	oService := ownershipService.NewOwnershipService(oRepo, oProductAdapter, oEmailAdapter, oUserAdapter, pub)
	oHandler := ownershipHandler.NewOwnershipHandler(oService)

	// Warranty Module
	wRepo := warrantyRepo.NewWarrantyRepository(database)
	wService := warrantyService.NewWarrantyService(wRepo, pub)
	wHandler := warrantyHandler.NewWarrantyHandler(wService)

	// Warranty Claim Module
	wcRepo := warrantyClaimRepo.NewWarrantyClaimRepository(database)
	wcService := warrantyClaimService.NewWarrantyClaimService(wcRepo, wRepo)
	wcHandler := warrantyClaimHandler.NewWarrantyClaimHandler(wcService)

	// AutoMigrate
	_ = database.AutoMigrate(&ownershipEntity.Ownership{}, &warrantyEntity.Warranty{}, &warrantyClaimEntity.WarrantyClaim{})

	// Initialize Public Module
	pubService := publicService.NewPublicService(productItemsRepo, tRepo, oRepo, wRepo, lRepo, uRepo)
	pubHandler := publicHandler.NewPublicHandler(pubService)

	r := router.SetupRouter(router.RouterDependency{
		BatchHandler:                 batchHandler,
		AuthHandler:                  aHandler,
		UserHandler:                  uHandler,
		ProductHandler:               pHandler,
		ProductCategoryHandler:       cHandler,
		AuditHandler:                 auditHandler,
		LocationHandler:              lHandler,
		DashboardHandler:             dbHandler,
		UserRepo:                     uRepo,
		ProductVariantHandler:        vHandler,
		ProductAttributeHandler:      pAttrHandler,
		ProductAttributeValueHandler: pAttrValHandler,
		TraceHandler:                 tHandler,
		PublicHandler:                pubHandler,
		RateLimiter:                  rateLimiter,
		ProductItemHandler:           piHandler,
		OwnershipHandler:             oHandler,
		WarrantyHandler:              wHandler,
		WarrantyClaimHandler:         wcHandler,
	})

	return &App{
		Router: r,
	}
}
