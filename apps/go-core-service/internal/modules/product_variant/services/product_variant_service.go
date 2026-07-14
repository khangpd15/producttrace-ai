package services

import (
    "context"
    "encoding/json"
    "errors"

    "github.com/google/uuid"
    attrValRepos "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/repositories"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/dto/request"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/entities"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/repositories"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/dbctx"
    "gorm.io/datatypes"
    "gorm.io/gorm"
)

type ProductVariantService interface {
    UpdateVariant(ctx context.Context, id uuid.UUID, req request.UpdateVariantRequest) (*entities.ProductVariant, error)
    GetVariantByID(ctx context.Context, id uuid.UUID) (*entities.ProductVariant, error)
    GetVariantsByProductID(ctx context.Context, productID uuid.UUID, filter request.ListVariantRequest) ([]entities.ProductVariant, int64, error)
    DeleteVariant(ctx context.Context, id uuid.UUID) error
}

type productVariantService struct {
    db          *gorm.DB
    variantRepo repositories.ProductVariantRepository
    attrValRepo attrValRepos.AttributeValueRepository
}

func NewProductVariantService(
    db *gorm.DB,
    variantRepo repositories.ProductVariantRepository,
    attrValRepo attrValRepos.AttributeValueRepository,
) ProductVariantService {
    return &productVariantService{db: db, variantRepo: variantRepo, attrValRepo: attrValRepo}
}

func (s *productVariantService) UpdateVariant(ctx context.Context, id uuid.UUID, req request.UpdateVariantRequest) (*entities.ProductVariant, error) {
    variant, err := s.variantRepo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperror.NewNotFound("variant")
        }
        return nil, apperror.Wrap(err, apperror.NewInternal("Failed to find variant"))
    }

    if req.SKU != nil {
        exists, err := s.variantRepo.ExistsBySKUExcludeID(ctx, *req.SKU, id)
        if err != nil {
            return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check SKU"))
        }
        if exists {
            return nil, apperror.NewConflict("SKU already exists")
        }
        variant.SKU = *req.SKU
    }

    if req.Barcode != nil {
        exists, err := s.variantRepo.ExistsByBarcodeExcludeID(ctx, *req.Barcode, id)
        if err != nil {
            return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check barcode"))
        }
        if exists {
            return nil, apperror.NewConflict("Barcode already exists")
        }
        variant.Barcode = req.Barcode
    }

    if req.Name != nil {
        variant.Name = *req.Name
    }
    if req.Price != nil {
        variant.Price = req.Price
    }
    if req.Currency != nil {
        variant.Currency = req.Currency
    }
    if req.Status != nil {
        variant.Status = req.Status
    }
    if req.Images != nil {
        imagesJSON, _ := json.Marshal(req.Images)
        variant.ImagesJSON = datatypes.JSON(imagesJSON)
    }

    if err := s.variantRepo.Update(ctx, variant); err != nil {
        return nil, apperror.Wrap(err, apperror.NewInternal("Failed to update variant"))
    }

    return variant, nil
}

func (s *productVariantService) GetVariantByID(ctx context.Context, id uuid.UUID) (*entities.ProductVariant, error) {
    variant, err := s.variantRepo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperror.NewNotFound("variant")
        }
        return nil, apperror.Wrap(err, apperror.NewInternal("Failed to find variant"))
    }
    return variant, nil
}

func (s *productVariantService) GetVariantsByProductID(ctx context.Context, productID uuid.UUID, filter request.ListVariantRequest) ([]entities.ProductVariant, int64, error) {
    variants, err := s.variantRepo.FindByProductID(ctx, productID)
    if err != nil {
        return nil, 0, apperror.Wrap(err, apperror.NewInternal("Failed to get variants"))
    }
    return variants, int64(len(variants)), nil
}

func (s *productVariantService) DeleteVariant(ctx context.Context, id uuid.UUID) error {
    _, err := s.variantRepo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return apperror.NewNotFound("variant")
        }
        return apperror.Wrap(err, apperror.NewInternal("Failed to find variant"))
    }

    // Xoá riêng 1 variant: chỉ cascade XUỐNG (attribute_values của variant này),
    // TUYỆT ĐỐI không đụng tới product cha. Cả 2 lệnh dưới đây dùng chung 1
    // transaction để tránh trường hợp variant bị xoá nhưng attribute_value
    // của nó vẫn còn sót lại (hoặc ngược lại).
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        txCtx := dbctx.InjectTx(ctx, tx)

        if err := s.variantRepo.SoftDelete(txCtx, id); err != nil {
            return apperror.Wrap(err, apperror.NewInternal("Failed to delete variant"))
        }

        if err := s.attrValRepo.DeleteByVariantID(txCtx, id); err != nil {
            return apperror.Wrap(err, apperror.NewInternal("Failed to delete variant attribute values"))
        }

        return nil
    })
}