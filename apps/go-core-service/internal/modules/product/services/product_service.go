package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	eventPublisher "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/publisher"
	eventTypes "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/types"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/repositories"
	attrRepos "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/repositories"
	attrValRequest "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/dto/request"
	attrValEntities "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/entities"
	attrValRepos "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/repositories"
	categoryRepos "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/repositories"
	variantEntities "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/entities"
	variantRepos "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/dbctx"
	pkgResponse "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ProductService interface {
	CreateProduct(ctx context.Context, req request.CreateProductRequest, createdBy uuid.UUID) (*entities.Product, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, req request.UpdateProductRequest) (*entities.Product, error)
	GetProductByID(ctx context.Context, id uuid.UUID) (*entities.Product, error)
	GetAllProducts(ctx context.Context, filter request.ListProductRequest) (*response.ProductListResponse, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) error
}

type productService struct {
	db           *gorm.DB
	productRepo  repositories.ProductRepository
	variantRepo  variantRepos.ProductVariantRepository
	attrValRepo  attrValRepos.AttributeValueRepository
	attrRepo     attrRepos.AttributeRepository
	categoryRepo categoryRepos.ProductCategoryRepository
	publisher    *eventPublisher.Publisher
}

func NewProductService(
	db *gorm.DB,
	productRepo repositories.ProductRepository,
	variantRepo variantRepos.ProductVariantRepository,
	attrValRepo attrValRepos.AttributeValueRepository,
	attrRepo attrRepos.AttributeRepository,
	categoryRepo categoryRepos.ProductCategoryRepository,
	publisher *eventPublisher.Publisher,
) ProductService {
	return &productService{
		db:           db,
		productRepo:  productRepo,
		variantRepo:  variantRepo,
		attrValRepo:  attrValRepo,
		attrRepo:     attrRepo,
		categoryRepo: categoryRepo,
		publisher:    publisher,
	}
}

func (s *productService) CreateProduct(ctx context.Context, req request.CreateProductRequest, createdBy uuid.UUID) (*entities.Product, error) {
	// Category phải tồn tại trước khi cho tạo product thuộc category đó,
	// tránh product trỏ tới category_id "ma" (không tồn tại hoặc đã bị xoá).
	categoryExists, err := s.categoryRepo.ExistsByID(ctx, req.CategoryID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check category"))
	}
	if !categoryExists {
		return nil, apperror.NewNotFound("category")
	}

	// Chặn trùng SKU/barcode NGAY TRONG cùng 1 request (2 variant mới trùng nhau)
	// trước khi đi check trùng với DB.
	seenSKUs := make(map[string]bool)
	seenBarcodes := make(map[string]bool)
	for _, v := range req.Variants {
		if seenSKUs[v.SKU] {
			return nil, apperror.NewBadRequest("Duplicate SKU in request: " + v.SKU)
		}
		seenSKUs[v.SKU] = true

		exists, err := s.variantRepo.ExistsBySKU(ctx, v.SKU)
		if err != nil {
			return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check SKU"))
		}
		if exists {
			return nil, apperror.NewConflict("SKU already exists: " + v.SKU)
		}

		if v.Barcode != nil {
			if seenBarcodes[*v.Barcode] {
				return nil, apperror.NewBadRequest("Duplicate barcode in request: " + *v.Barcode)
			}
			seenBarcodes[*v.Barcode] = true

			exists, err = s.variantRepo.ExistsByBarcode(ctx, *v.Barcode)
			if err != nil {
				return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check barcode"))
			}
			if exists {
				return nil, apperror.NewConflict("Barcode already exists: " + *v.Barcode)
			}
		}
	}

	// Convert tags và metadata sang JSON
	tagsJSON, _ := json.Marshal(req.Tags)
	metadataJSON, _ := json.Marshal(req.Metadata)

	var product entities.Product
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		product = entities.Product{
			ID:           uuid.New(),
			CategoryID:   &req.CategoryID,
			Name:         req.Name,
			Slug:         &req.Slug,
			Description:  req.Description,
			ThumbnailURL: req.ThumbnailURL,
			Tags:         datatypes.JSON(tagsJSON),
			MetadataJSON: datatypes.JSON(metadataJSON),
			Status:       &req.Status,
			CreatedBy:    &createdBy,
		}
		txCtx := dbctx.InjectTx(ctx, tx)
		if err := s.productRepo.Create(txCtx, &product); err != nil {
			return apperror.Wrap(err, apperror.NewInternal("Failed to create product"))
		}

		for _, v := range req.Variants {
			imagesJSON, _ := json.Marshal(v.Images)
			variant := variantEntities.ProductVariant{
				ID:         uuid.New(),
				ProductID:  product.ID,
				SKU:        v.SKU,
				Name:       v.Name,
				Barcode:    v.Barcode,
				Price:      v.Price,
				Currency:   v.Currency,
				ImagesJSON: datatypes.JSON(imagesJSON),
				Status:     v.Status,
			}
			if err := s.variantRepo.Create(txCtx, &variant); err != nil {
				return apperror.Wrap(err, apperror.NewInternal("Failed to create variant"))
			}

			// product -> variant -> attributes: tạo attribute value ngay
			// trong cùng transaction với product/variant ở trên.
			if err := s.createVariantAttributes(txCtx, variant.ID, v.Attributes, &req.CategoryID, v.SKU); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	event := eventTypes.Event{
		EventID:       uuid.New().String(),
		EventType:     "product.created",
		EventVersion:  "1.0",
		Timestamp:     time.Now(),
		Producer:      "go-core-service",
		CorrelationID: product.ID.String(),
		Payload: map[string]interface{}{
			"product_id":  product.ID.String(),
			"name":        product.Name,
			"category_id": product.CategoryID,
		},
	}

	if err := s.publisher.Publish(event); err != nil {
		log.Printf("failed to publish event: %v", err)
	}

	return s.loadProductDetail(ctx, product.ID)
}

func (s *productService) UpdateProduct(ctx context.Context, id uuid.UUID, req request.UpdateProductRequest) (*entities.Product, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound("product")
		}
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to find product"))
	}

	// Category hiệu lực SAU khi áp dụng thay đổi (nếu có) - dùng để validate
	// attribute có thuộc đúng category không khi tạo/sửa attribute value bên dưới.
	effectiveCategoryID := product.CategoryID
	if req.CategoryID != nil {
		categoryExists, err := s.categoryRepo.ExistsByID(ctx, *req.CategoryID)
		if err != nil {
			return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check category"))
		}
		if !categoryExists {
			return nil, apperror.NewNotFound("category")
		}
		effectiveCategoryID = req.CategoryID
	}

	// Chặn trùng SKU/barcode NGAY TRONG cùng 1 request trước khi vào transaction.
	seenSKUs := make(map[string]bool)
	seenBarcodes := make(map[string]bool)
	for _, v := range req.Variants {
		if v.SKU != nil {
			if seenSKUs[*v.SKU] {
				return nil, apperror.NewBadRequest("Duplicate SKU in request: " + *v.SKU)
			}
			seenSKUs[*v.SKU] = true
		}
		if v.Barcode != nil {
			if seenBarcodes[*v.Barcode] {
				return nil, apperror.NewBadRequest("Duplicate barcode in request: " + *v.Barcode)
			}
			seenBarcodes[*v.Barcode] = true
		}
		// Variant mới (không có ID) bắt buộc phải có SKU và Name.
		if v.ID == nil {
			if v.SKU == nil || *v.SKU == "" {
				return nil, apperror.NewBadRequest("New variant requires sku")
			}
			if v.Name == nil || *v.Name == "" {
				return nil, apperror.NewBadRequest("New variant requires name")
			}
		}
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := dbctx.InjectTx(ctx, tx)

		if req.Name != nil {
			product.Name = *req.Name
		}
		if req.CategoryID != nil {
			product.CategoryID = req.CategoryID
		}
		if req.Description != nil {
			product.Description = req.Description
		}
		if req.Slug != nil {
			product.Slug = req.Slug
		}
		if req.ThumbnailURL != nil {
			product.ThumbnailURL = req.ThumbnailURL
		}
		if req.Status != nil {
			product.Status = req.Status
		}
		if req.Tags != nil {
			tagsJSON, _ := json.Marshal(req.Tags)
			product.Tags = datatypes.JSON(tagsJSON)
		}
		if req.Metadata != nil {
			metadataJSON, _ := json.Marshal(req.Metadata)
			product.MetadataJSON = datatypes.JSON(metadataJSON)
		}

		if err := s.productRepo.Update(txCtx, product); err != nil {
			return apperror.Wrap(err, apperror.NewInternal("Failed to update product"))
		}

		// req.Variants == nil (FE không gửi key "variants") => giữ nguyên
		// toàn bộ variant hiện có, không đụng gì tới chúng.
		if req.Variants == nil {
			return nil
		}

		existingVariants, err := s.variantRepo.FindByProductID(txCtx, id)
		if err != nil {
			return apperror.Wrap(err, apperror.NewInternal("Failed to load existing variants"))
		}
		existingByID := make(map[uuid.UUID]variantEntities.ProductVariant, len(existingVariants))
		for _, ev := range existingVariants {
			existingByID[ev.ID] = ev
		}

		keptIDs := make(map[uuid.UUID]bool)

		for _, v := range req.Variants {
			if v.ID != nil {
				// --- Cập nhật 1 variant đã tồn tại ---
				existing, ok := existingByID[*v.ID]
				if !ok {
					return apperror.NewNotFound("variant: " + v.ID.String())
				}
				keptIDs[*v.ID] = true

				if v.SKU != nil {
					exists, err := s.variantRepo.ExistsBySKUExcludeID(txCtx, *v.SKU, existing.ID)
					if err != nil {
						return apperror.Wrap(err, apperror.NewInternal("Failed to check SKU"))
					}
					if exists {
						return apperror.NewConflict("SKU already exists: " + *v.SKU)
					}
					existing.SKU = *v.SKU
				}
				if v.Barcode != nil {
					exists, err := s.variantRepo.ExistsByBarcodeExcludeID(txCtx, *v.Barcode, existing.ID)
					if err != nil {
						return apperror.Wrap(err, apperror.NewInternal("Failed to check barcode"))
					}
					if exists {
						return apperror.NewConflict("Barcode already exists: " + *v.Barcode)
					}
					existing.Barcode = v.Barcode
				}
				if v.Name != nil {
					existing.Name = *v.Name
				}
				if v.Price != nil {
					existing.Price = v.Price
				}
				if v.Currency != nil {
					existing.Currency = v.Currency
				}
				if v.Status != nil {
					existing.Status = v.Status
				}
				if v.Images != nil {
					imagesJSON, _ := json.Marshal(v.Images)
					existing.ImagesJSON = datatypes.JSON(imagesJSON)
				}

				if err := s.variantRepo.Update(txCtx, &existing); err != nil {
					return apperror.Wrap(err, apperror.NewInternal("Failed to update variant: "+existing.SKU))
				}

				// v.Attributes == nil => không đổi gì cả.
				// v.Attributes != nil (kể cả rỗng) => thay thế toàn bộ attribute value hiện có.
				if v.Attributes != nil {
					if err := s.attrValRepo.DeleteByVariantID(txCtx, existing.ID); err != nil {
						return apperror.Wrap(err, apperror.NewInternal("Failed to clear old attribute values for variant: "+existing.SKU))
					}
					if err := s.createVariantAttributes(txCtx, existing.ID, *v.Attributes, effectiveCategoryID, existing.SKU); err != nil {
						return err
					}
				}
			} else {
				// --- Tạo mới 1 variant trong lúc sửa product ---
				exists, err := s.variantRepo.ExistsBySKU(txCtx, *v.SKU)
				if err != nil {
					return apperror.Wrap(err, apperror.NewInternal("Failed to check SKU"))
				}
				if exists {
					return apperror.NewConflict("SKU already exists: " + *v.SKU)
				}
				if v.Barcode != nil {
					exists, err = s.variantRepo.ExistsByBarcode(txCtx, *v.Barcode)
					if err != nil {
						return apperror.Wrap(err, apperror.NewInternal("Failed to check barcode"))
					}
					if exists {
						return apperror.NewConflict("Barcode already exists: " + *v.Barcode)
					}
				}

				imagesJSON, _ := json.Marshal(v.Images)
				newVariant := variantEntities.ProductVariant{
					ID:         uuid.New(),
					ProductID:  id,
					SKU:        *v.SKU,
					Name:       *v.Name,
					Barcode:    v.Barcode,
					Price:      v.Price,
					Currency:   v.Currency,
					ImagesJSON: datatypes.JSON(imagesJSON),
					Status:     v.Status,
				}
				if err := s.variantRepo.Create(txCtx, &newVariant); err != nil {
					return apperror.Wrap(err, apperror.NewInternal("Failed to create variant: "+*v.SKU))
				}
				keptIDs[newVariant.ID] = true

				if v.Attributes != nil {
					if err := s.createVariantAttributes(txCtx, newVariant.ID, *v.Attributes, effectiveCategoryID, *v.SKU); err != nil {
						return err
					}
				}
			}
		}

		// Variant nào đang tồn tại nhưng KHÔNG có mặt trong req.Variants nữa
		// => FE đã xoá variant đó trên form => soft-delete + cascade xoá
		// attribute value của nó, trong cùng transaction.
		for _, ev := range existingVariants {
			if keptIDs[ev.ID] {
				continue
			}
			if err := s.variantRepo.SoftDelete(txCtx, ev.ID); err != nil {
				return apperror.Wrap(err, apperror.NewInternal("Failed to delete removed variant: "+ev.SKU))
			}
			if err := s.attrValRepo.DeleteByVariantID(txCtx, ev.ID); err != nil {
				return apperror.Wrap(err, apperror.NewInternal("Failed to delete attribute values of removed variant: "+ev.SKU))
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.loadProductDetail(ctx, id)
}

// createVariantAttributes validate rồi tạo 1 danh sách attribute value cho
// 1 variant. Dùng chung cho cả CreateProduct, tạo variant mới trong lúc sửa
// product, và thay thế attribute value khi sửa 1 variant đã tồn tại.
func (s *productService) createVariantAttributes(
	ctx context.Context,
	variantID uuid.UUID,
	items []attrValRequest.CreateAttributeValueRequest,
	categoryID *uuid.UUID,
	variantLabel string,
) error {
	if len(items) == 0 {
		return nil
	}
	if categoryID == nil {
		return apperror.NewBadRequest("Product has no category, cannot assign attributes")
	}

	seenAttributes := make(map[uuid.UUID]bool)
	for _, attr := range items {
		if seenAttributes[attr.AttributeID] {
			return apperror.NewBadRequest("Duplicate attribute assignment for variant SKU: " + variantLabel)
		}
		seenAttributes[attr.AttributeID] = true

		attrEntity, err := s.attrRepo.FindByID(ctx, attr.AttributeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperror.NewNotFound("attribute")
			}
			return apperror.Wrap(err, apperror.NewInternal("Failed to check attribute"))
		}

		// Attribute phải thuộc đúng category của product, tránh gán nhầm
		// attribute của category khác (vd: "Dung lượng pin" của Điện thoại
		// bị gán cho variant thuộc category "Giày dép").
		if attrEntity.CategoryID != *categoryID {
			return apperror.NewBadRequest("Attribute does not belong to this product's category: " + attr.AttributeID.String())
		}

		val := attrValEntities.AttributeValue{
			ID:               uuid.New(),
			ProductVariantID: variantID,
			AttributeID:      attr.AttributeID,
			Label:            attr.Label,
			ValueText:        attr.ValueText,
			ValueNumber:      attr.ValueNumber,
			ValueBoolean:     attr.ValueBoolean,
			CreatedAt:        time.Now(),
		}
		if err := s.attrValRepo.Create(ctx, &val); err != nil {
			return apperror.Wrap(err, apperror.NewInternal("Failed to create attribute value for variant SKU: "+variantLabel))
		}
	}
	return nil
}

// loadProductDetail load lại product kèm Variants (qua Preload có sẵn ở
// productRepo.FindByID) rồi tự tay gắn AttributeValues cho từng variant
// (field transient, không qua GORM Preload) trước khi trả về cho mapper.
func (s *productService) loadProductDetail(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound("product")
		}
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to load product"))
	}
	if err := s.populateVariantAttributeValues(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

func (s *productService) populateVariantAttributeValues(ctx context.Context, product *entities.Product) error {
	for i := range product.Variants {
		vals, err := s.attrValRepo.FindByVariantID(ctx, product.Variants[i].ID)
		if err != nil {
			return apperror.Wrap(err, apperror.NewInternal("Failed to load variant attributes"))
		}
		product.Variants[i].AttributeValues = vals
	}
	return nil
}

func (s *productService) GetProductByID(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	return s.loadProductDetail(ctx, id)
}

func (s *productService) GetAllProducts(ctx context.Context, filter request.ListProductRequest) (*response.ProductListResponse, error) {
	var categoryID *uuid.UUID
	if filter.CategoryID != nil {
		id, err := uuid.Parse(*filter.CategoryID)
		if err != nil {
			return nil, apperror.NewBadRequest("Invalid category_id")
		}
		categoryID = &id
	}

	repoFilter := repositories.ProductFilter{
		Search:     filter.Search,
		CategoryID: categoryID,
		Status:     filter.Status,
		Page:       filter.Page,
		Limit:      filter.Limit,
		SortBy:     filter.SortBy,
		SortOrder:  filter.SortOrder,
	}

	items, total, err := s.productRepo.FindAllWithStats(ctx, repoFilter)
	if err != nil {
		return nil, apperror.WrapDBError(err, "products")
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages == 0 && total > 0 {
		totalPages = 1
	}

	meta := pkgResponse.PaginationMeta{
		CurrentPage: page,
		PageSize:    limit,
		TotalItems:  int(total),
		TotalPages:  totalPages,
	}

	return &response.ProductListResponse{
		Items: items,
		Meta:  meta,
	}, nil
}

func (s *productService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	_, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NewNotFound("product")
		}
		return apperror.Wrap(err, apperror.NewInternal("Failed to find product"))
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := dbctx.InjectTx(ctx, tx)

		// Xoá theo đúng thứ tự cha -> con: product -> variant -> attribute_value.
		// Cả 3 lệnh dùng chung 1 transaction (txCtx) nên hoặc cùng thành công,
		// hoặc cùng rollback -> không còn tình trạng xoá dở (product mất
		// nhưng variant/attribute_value còn, hay ngược lại).
		if err := s.productRepo.SoftDelete(txCtx, id); err != nil {
			return apperror.Wrap(err, apperror.NewInternal("Failed to delete product"))
		}

		if err := s.variantRepo.SoftDeleteByProductID(txCtx, id); err != nil {
			return apperror.Wrap(err, apperror.NewInternal("Failed to delete product variants"))
		}

		// attribute_values không có cột is_deleted trong schema (xem
		// migrations/000001) nên xoá cứng (hard delete) là đúng ở đây.
		if err := s.attrValRepo.DeleteByProductID(txCtx, id); err != nil {
			return apperror.Wrap(err, apperror.NewInternal("Failed to delete variant attribute values"))
		}

		return nil
	})
}
