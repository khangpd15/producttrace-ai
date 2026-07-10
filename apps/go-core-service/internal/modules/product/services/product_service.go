package services

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/repositories"
	variantEntities "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/entities"
	variantRepos "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
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
	db          *gorm.DB
	productRepo repositories.ProductRepository
	variantRepo variantRepos.ProductVariantRepository
}

func NewProductService(
	db *gorm.DB,
	productRepo repositories.ProductRepository,
	variantRepo variantRepos.ProductVariantRepository,
) ProductService {
	return &productService{
		db:          db,
		productRepo: productRepo,
		variantRepo: variantRepo,
	}
}

func (s *productService) CreateProduct(ctx context.Context, req request.CreateProductRequest, createdBy uuid.UUID) (*entities.Product, error) {
	for _, v := range req.Variants {
		exists, err := s.variantRepo.ExistsBySKU(ctx, v.SKU)
		if err != nil {
			return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check SKU"))
		}
		if exists {
			return nil, apperror.NewConflict("SKU already exists: " + v.SKU)
		}

		if v.Barcode != nil {
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
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		txCtx := variantRepos.InjectTx(ctx, tx)
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
			}
			if err := s.variantRepo.Create(txCtx, &variant); err != nil {
				return apperror.Wrap(err, apperror.NewInternal("Failed to create variant"))
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *productService) UpdateProduct(ctx context.Context, id uuid.UUID, req request.UpdateProductRequest) (*entities.Product, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound("product")
		}
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to find product"))
	}

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

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to update product"))
	}

	return product, nil
}

func (s *productService) GetProductByID(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound("product")
		}
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to find product"))
	}
	return product, nil
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

	if err := s.productRepo.SoftDelete(ctx, id); err != nil {
		return apperror.Wrap(err, apperror.NewInternal("Failed to delete product"))
	}

	return nil
}
