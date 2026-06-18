package services

import (
    "context"
    "errors"

    "github.com/google/uuid"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/dto"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/entities"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/repositories"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
    "gorm.io/gorm"
)

type ProductService interface {
    CreateProduct(ctx context.Context, req dto.CreateProductRequest, createdBy uuid.UUID) (*entities.Product, error)
    UpdateProduct(ctx context.Context, id uuid.UUID, req dto.UpdateProductRequest) (*entities.Product, error)
    GetProductByID(ctx context.Context, id uuid.UUID) (*entities.Product, error)
    GetAllProducts(ctx context.Context, filter dto.ListProductRequest) ([]entities.Product, int64, error)
    DeleteProduct(ctx context.Context, id uuid.UUID) error
}

type productService struct {
    db          *gorm.DB
    productRepo repositories.ProductRepository
    variantRepo repositories.ProductVariantRepository
}

func NewProductService(
    db *gorm.DB,
    productRepo repositories.ProductRepository,
    variantRepo repositories.ProductVariantRepository,
) ProductService {
    return &productService{
        db:          db,
        productRepo: productRepo,
        variantRepo: variantRepo,
    }
}

func (s *productService) CreateProduct(ctx context.Context, req dto.CreateProductRequest, createdBy uuid.UUID) (*entities.Product, error) {
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

    var product entities.Product
    err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        product = entities.Product{
            ID:           uuid.New(),
            CategoryID:   &req.CategoryID,
            Name:         req.Name,
            Description:  req.Description,
            ThumbnailURL: req.ThumbnailURL,
            Status:       &req.Status,
            CreatedBy:    &createdBy,
        }
        txCtx := repositories.InjectTx(ctx, tx)
        if err := s.productRepo.Create(txCtx, &product); err != nil {
            return apperror.Wrap(err, apperror.NewInternal("Failed to create product"))
        }

        for _, v := range req.Variants {
            variant := entities.ProductVariant{
                ID:        uuid.New(),
                ProductID: product.ID,
                SKU:       v.SKU,
                Name:      v.Name,
                Barcode:   v.Barcode,
                Price:     v.Price,
                Currency:  v.Currency,
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

func (s *productService) UpdateProduct(ctx context.Context, id uuid.UUID, req dto.UpdateProductRequest) (*entities.Product, error) {
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
    if req.ThumbnailURL != nil {
        product.ThumbnailURL = req.ThumbnailURL
    }
    if req.Status != nil {
        product.Status = req.Status
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

func (s *productService) GetAllProducts(ctx context.Context, filter dto.ListProductRequest) ([]entities.Product, int64, error) {
    var categoryID *uuid.UUID
    if filter.CategoryID != nil {
        id, err := uuid.Parse(*filter.CategoryID)
        if err != nil {
            return nil, 0, apperror.NewBadRequest("Invalid category_id")
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

    return s.productRepo.FindAll(ctx, repoFilter)
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