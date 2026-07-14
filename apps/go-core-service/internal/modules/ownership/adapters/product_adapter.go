package adapters

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/service"
	"gorm.io/gorm"
)

type ProductAdapter struct {
	db *gorm.DB
}

func NewProductAdapter(db *gorm.DB) service.IProductService {
	return &ProductAdapter{db: db}
}

func (a *ProductAdapter) FindProductByQR(ctx context.Context, qrCode string) (uuid.UUID, error) {
	var item struct {
		ID uuid.UUID
	}
	query := a.db.WithContext(ctx).Table("product_items")
	if _, err := uuid.Parse(qrCode); err == nil {
		query = query.Where("(id = ? OR item_code = ? OR serial_number = ?) AND is_deleted = false", qrCode, qrCode, qrCode)
	} else {
		query = query.Where("(item_code = ? OR serial_number = ?) AND is_deleted = false", qrCode, qrCode)
	}
	err := query.Select("id").First(&item).Error
	if err != nil {
		return uuid.Nil, err
	}
	return item.ID, nil
}

func (a *ProductAdapter) ValidateProductOwnershipStatus(ctx context.Context, productID uuid.UUID) error {
	var item struct {
		Status    string
		IsDeleted bool
	}
	err := a.db.WithContext(ctx).Table("product_items").
		Select("status, is_deleted").
		Where("id = ?", productID).
		First(&item).Error
	if err != nil {
		return err
	}
	if item.IsDeleted {
		return errors.New("product item is deleted")
	}
	if item.Status == "RECALLED" {
		return errors.New("product item is recalled")
	}
	if item.Status == "DAMAGED" {
		return errors.New("product item is damaged")
	}

	// Check if there is an active ownership record
	var count int64
	err = a.db.WithContext(ctx).Table("ownerships").
		Where("product_item_id = ? AND status = ?", productID, "ACTIVE").
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("product item is already registered/owned")
	}

	return nil
}

func (a *ProductAdapter) UpdateOwnershipStatus(ctx context.Context, productID uuid.UUID, status string) error {
	// Map "ACTIVE" status to "REGISTERED" status due to constraint chk_product_items_status
	dbStatus := status
	if status == "ACTIVE" {
		dbStatus = "REGISTERED"
	}

	// Also set RegisteredAt if status becomes REGISTERED
	updates := map[string]interface{}{
		"status":     dbStatus,
		"updated_at": time.Now(),
	}
	if dbStatus == "REGISTERED" {
		updates["registered_at"] = time.Now()
	}

	err := a.db.WithContext(ctx).Table("product_items").
		Where("id = ?", productID).
		Updates(updates).Error
	return err
}

func (a *ProductAdapter) GetProductItemDetail(ctx context.Context, productItemID uuid.UUID) (name string, sku string, err error) {
	var detail struct {
		ProductName string
		Sku         string
	}
	err = a.db.WithContext(ctx).Table("product_items pi").
		Select("p.name as product_name, pv.sku").
		Joins("JOIN product_variants pv ON pi.variant_id = pv.id").
		Joins("JOIN products p ON pv.product_id = p.id").
		Where("pi.id = ?", productItemID).
		First(&detail).Error
	if err != nil {
		return "", "", err
	}
	return detail.ProductName, detail.Sku, nil
}

func (a *ProductAdapter) SearchProductItemIDs(ctx context.Context, productName string, productCode string) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	query := a.db.WithContext(ctx).Table("product_items pi").
		Select("pi.id").
		Joins("JOIN product_variants pv ON pi.variant_id = pv.id").
		Joins("JOIN products p ON pv.product_id = p.id").
		Where("pi.is_deleted = false")

	if productName != "" {
		query = query.Where("p.name ILIKE ?", "%"+productName+"%")
	}
	if productCode != "" {
		query = query.Where("pi.item_code = ? OR pi.serial_number = ?", productCode, productCode)
	}

	err := query.Find(&ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}
