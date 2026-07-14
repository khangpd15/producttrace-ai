package adapters

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DbOwnershipAdapter struct {
	db *gorm.DB
}

func NewDbOwnershipAdapter(db *gorm.DB) OwnershipPort {
	return &DbOwnershipAdapter{db: db}
}

func (a *DbOwnershipAdapter) VerifyOwnership(ctx context.Context, userID uuid.UUID, productItemID uuid.UUID) (bool, error) {
	var count int64
	err := a.db.WithContext(ctx).Table("ownerships").
		Where("owner_id = ? AND product_item_id = ? AND status = ?", userID, productItemID, "ACTIVE").
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (a *DbOwnershipAdapter) GetActiveOwner(ctx context.Context, productItemID uuid.UUID) (uuid.UUID, error) {
	var own struct {
		OwnerID uuid.UUID
	}
	err := a.db.WithContext(ctx).Table("ownerships").
		Select("owner_id").
		Where("product_item_id = ? AND status = ?", productItemID, "ACTIVE").
		First(&own).Error
	if err != nil {
		return uuid.Nil, err
	}
	return own.OwnerID, nil
}

type DbProductItemAdapter struct {
	db *gorm.DB
}

func NewDbProductItemAdapter(db *gorm.DB) ProductItemPort {
	return &DbProductItemAdapter{db: db}
}

func (a *DbProductItemAdapter) CheckWarrantyValidity(ctx context.Context, productItemID uuid.UUID) (bool, error) {
	// First check if the product item has an active ownership record
	var hasActiveOwnership int64
	err := a.db.WithContext(ctx).Table("ownerships").
		Where("product_item_id = ? AND status = ?", productItemID, "ACTIVE").
		Count(&hasActiveOwnership).Error
	if err != nil {
		return false, err
	}
	if hasActiveOwnership == 0 {
		return false, nil
	}

	// Then check if the product item is recalled or deleted
	var item struct {
		Status    string
		IsDeleted bool
	}
	err = a.db.WithContext(ctx).Table("product_items").
		Select("status, is_deleted").
		Where("id = ?", productItemID).
		First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	if item.IsDeleted || item.Status == "RECALLED" || item.Status == "DAMAGED" {
		return false, nil
	}

	// Check warranties table if a record exists
	var warranty struct {
		Status  string
		EndDate time.Time
	}
	err = a.db.WithContext(ctx).Table("warranties").
		Select("status, end_date").
		Where("product_item_id = ?", productItemID).
		Order("created_at DESC").
		First(&warranty).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Fallback: valid if registered_at is not zero and registered_at + 2 years is in the future
			var itemReg struct {
				RegisteredAt *time.Time
			}
			err = a.db.WithContext(ctx).Table("product_items").
				Select("registered_at").
				Where("id = ?", productItemID).
				First(&itemReg).Error
			if err != nil {
				return false, err
			}
			if itemReg.RegisteredAt == nil {
				return false, nil
			}
			expireDate := itemReg.RegisteredAt.AddDate(2, 0, 0)
			return expireDate.After(time.Now()), nil
		}
		return false, err
	}

	if warranty.Status == "EXPIRED" || warranty.Status == "CANCELLED" {
		return false, nil
	}
	if !warranty.EndDate.IsZero() && warranty.EndDate.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}
