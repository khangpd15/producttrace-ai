package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/entity"
	"gorm.io/gorm"
)

type ActiveOwnershipInfo struct {
	ProductItemID uuid.UUID `gorm:"column:product_item_id"`
	OwnerID       uuid.UUID `gorm:"column:owner_id"`
	OwnerName     string    `gorm:"column:full_name"`
	OwnerEmail    string    `gorm:"column:email"`
	OwnedAt       time.Time `gorm:"column:owned_at"`
}

type WarrantyRepository interface {
	Create(ctx context.Context, warranty *entity.Warranty) error
	Update(ctx context.Context, warranty *entity.Warranty) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Warranty, error)
	FindByWarrantyCode(ctx context.Context, code string) (*entity.Warranty, error)
	FindBySerialNumber(ctx context.Context, serialNumber string) ([]entity.Warranty, error)
	FindAll(ctx context.Context) ([]entity.Warranty, error)
	GetOwnershipDate(ctx context.Context, itemCode, serialNumber string) (*time.Time, error)
	GetActiveOwnership(ctx context.Context, itemCode, serialNumber string) (*ActiveOwnershipInfo, error)
	FindByProductItemID(ctx context.Context, productItemID uuid.UUID) (*entity.Warranty, error)
	FindMyWarranties(ctx context.Context, ownerID uuid.UUID) ([]entity.Warranty, error)
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

type gormWarrantyRepository struct {
	db *gorm.DB
}

func NewWarrantyRepository(db *gorm.DB) WarrantyRepository {
	return &gormWarrantyRepository{db: db}
}

func (r *gormWarrantyRepository) Create(ctx context.Context, warranty *entity.Warranty) error {
	return r.db.WithContext(ctx).Create(warranty).Error
}

func (r *gormWarrantyRepository) Update(ctx context.Context, warranty *entity.Warranty) error {
	return r.db.WithContext(ctx).Save(warranty).Error
}

func (r *gormWarrantyRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Warranty, error) {
	var warranty entity.Warranty
	if err := r.db.WithContext(ctx).First(&warranty, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &warranty, nil
}

func (r *gormWarrantyRepository) FindByWarrantyCode(ctx context.Context, code string) (*entity.Warranty, error) {
	var warranty entity.Warranty
	if err := r.db.WithContext(ctx).First(&warranty, "warranty_code = ?", code).Error; err != nil {
		return nil, err
	}
	return &warranty, nil
}

func (r *gormWarrantyRepository) FindBySerialNumber(ctx context.Context, serialNumber string) ([]entity.Warranty, error) {
	var warranties []entity.Warranty
	if err := r.db.WithContext(ctx).Where("serial_number = ?", serialNumber).Find(&warranties).Error; err != nil {
		return nil, err
	}
	return warranties, nil
}

func (r *gormWarrantyRepository) FindAll(ctx context.Context) ([]entity.Warranty, error) {
	var warranties []entity.Warranty
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&warranties).Error; err != nil {
		return nil, err
	}
	return warranties, nil
}

func (r *gormWarrantyRepository) GetOwnershipDate(ctx context.Context, itemCode, serialNumber string) (*time.Time, error) {
	// Use Scan with GORM result to detect zero rows and return gorm.ErrRecordNotFound
	var res struct {
		OwnedAt time.Time `gorm:"column:owned_at"`
	}

	tx := r.db.WithContext(ctx).
		Table("ownerships").
		Select("ownerships.owned_at").
		Joins("JOIN product_items ON product_items.id = ownerships.product_item_id").
		Where("product_items.item_code = ? AND product_items.serial_number = ? AND ownerships.status = ?", itemCode, serialNumber, "ACTIVE").
		Order("ownerships.owned_at DESC").
		Limit(1).
		Scan(&res)

	if tx.Error != nil {
		return nil, tx.Error
	}

	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &res.OwnedAt, nil
}

func (r *gormWarrantyRepository) GetActiveOwnership(ctx context.Context, itemCode, serialNumber string) (*ActiveOwnershipInfo, error) {
	var res ActiveOwnershipInfo
	tx := r.db.WithContext(ctx).
		Table("ownerships").
		Select("ownerships.product_item_id, ownerships.owner_id, ownerships.owned_at, users.full_name, users.email").
		Joins("JOIN product_items ON product_items.id = ownerships.product_item_id").
		Joins("JOIN users ON users.id = ownerships.owner_id").
		Where("product_items.item_code = ? AND product_items.serial_number = ? AND ownerships.status = ?", itemCode, serialNumber, "ACTIVE").
		Order("ownerships.owned_at DESC").
		Limit(1).
		Scan(&res)

	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &res, nil
}

func (r *gormWarrantyRepository) FindByProductItemID(ctx context.Context, productItemID uuid.UUID) (*entity.Warranty, error) {
	var warranty entity.Warranty
	if err := r.db.WithContext(ctx).First(&warranty, "product_item_id = ?", productItemID).Error; err != nil {
		return nil, err
	}
	return &warranty, nil
}

func (r *gormWarrantyRepository) FindMyWarranties(ctx context.Context, ownerID uuid.UUID) ([]entity.Warranty, error) {
	var warranties []entity.Warranty
	if err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Order("created_at desc").Find(&warranties).Error; err != nil {
		return nil, err
	}
	return warranties, nil
}

func (r *gormWarrantyRepository) Transaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}
