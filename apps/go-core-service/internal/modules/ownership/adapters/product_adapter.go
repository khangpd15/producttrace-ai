package adapters

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/entity"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/service"
	productRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/repositories"
	productItemsRepo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"gorm.io/gorm"
)

type RealProductAdapter struct {
	db              *gorm.DB
	productItemRepo productItemsRepo.ProductItemRepository
	productRepo     productRepo.ProductRepository
}

func NewRealProductAdapter(db *gorm.DB, productItemRepo productItemsRepo.ProductItemRepository, productRepo productRepo.ProductRepository) service.IProductService {
	return &RealProductAdapter{
		db:              db,
		productItemRepo: productItemRepo,
		productRepo:     productRepo,
	}
}

func (a *RealProductAdapter) FindProductByQR(ctx context.Context, qrCode string) (uuid.UUID, error) {
	var dest struct {
		ID uuid.UUID
	}
	err := a.db.WithContext(ctx).Table("product_items").
		Select("id").
		Where("is_deleted = false AND (item_code = ? OR serial_number = ? OR verification_token = ?)", qrCode, qrCode, qrCode).
		First(&dest).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return uuid.Nil, apperror.NewBadRequest("Không tìm thấy sản phẩm với mã QR / Item Code này")
		}
		return uuid.Nil, err
	}
	return dest.ID, nil
}

func (a *RealProductAdapter) ValidateProductOwnershipStatus(ctx context.Context, productID uuid.UUID) error {
	var item struct {
		Status string
	}
	err := entity.GetTx(ctx, a.db).WithContext(ctx).Table("product_items").
		Select("status").
		Where("id = ? AND is_deleted = false", productID).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NewNotFound("Sản phẩm")
		}
		return err
	}
	status := item.Status
	if status == "REGISTERED" || status == "WARRANTY_ACTIVE" || status == "WARRANTY_CLAIMED" {
		return apperror.NewBadRequest("Sản phẩm đã được đăng ký sở hữu hoặc đang trong thời gian bảo hành.")
	}
	return nil
}

func (a *RealProductAdapter) UpdateOwnershipStatus(ctx context.Context, productID uuid.UUID, status string) error {
	db := entity.GetTx(ctx, a.db)

	var dbStatus string
	switch status {
	case "ACTIVE":
		dbStatus = "REGISTERED"
	case "PENDING":
		return nil
	default:
		dbStatus = status
	}

	err := db.WithContext(ctx).Table("product_items").
		Where("id = ? AND is_deleted = false", productID).
		Update("status", dbStatus).Error
	if err != nil {
		return err
	}
	return nil
}

func (a *RealProductAdapter) GetProductItemDetail(ctx context.Context, productItemID uuid.UUID) (string, string, string, error) {
	item, err := a.productItemRepo.FindByID(ctx, productItemID)
	if err != nil {
		return "", "", "", err
	}

	sku := item.Variant.SKU
	productName := item.Variant.Name
	if item.Variant.ProductID != uuid.Nil {
		product, err := a.productRepo.FindByID(ctx, item.Variant.ProductID)
		if err == nil && product != nil {
			productName = product.Name
		}
	}

	if productName == "" {
		productName = item.ItemCode
	}

	return productName, sku, item.SerialNumber, nil
}

func (a *RealProductAdapter) SearchProductItemIDs(ctx context.Context, productName string, productCode string) ([]uuid.UUID, error) {
	query := a.db.WithContext(ctx).Table("product_items").
		Select("product_items.id").
		Joins("LEFT JOIN product_variants pv ON pv.id = product_items.variant_id").
		Joins("LEFT JOIN products p ON p.id = pv.product_id").
		Where("product_items.is_deleted = false")

	if strings.TrimSpace(productName) != "" {
		query = query.Where("p.name ILIKE ?", "%"+strings.TrimSpace(productName)+"%")
	}
	if strings.TrimSpace(productCode) != "" {
		code := "%" + strings.TrimSpace(productCode) + "%"
		query = query.Where("pv.sku ILIKE ? OR product_items.item_code ILIKE ? OR product_items.serial_number ILIKE ?", code, code, code)
	}

	var ids []uuid.UUID
	if err := query.Pluck("product_items.id", &ids).Error; err != nil {
		return nil, err
	}

	return ids, nil
}
