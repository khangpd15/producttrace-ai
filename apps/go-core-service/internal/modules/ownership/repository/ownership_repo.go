package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/entity"
	"gorm.io/gorm"
)

type IOwnershipRepository interface {
	CreateOwnership(ctx context.Context, tx *gorm.DB, ownership *entity.Ownership) (*entity.Ownership, error)
	// GetOwnershipByProductItemID trả về record sở hữu hiện tại (mới nhất)
	GetOwnershipByProductItemID(ctx context.Context, productItemID uuid.UUID) (*entity.Ownership, error)
	// GetOwnershipHistoryByProductItemID trả về toàn bộ lịch sử sở hữu theo thứ tự thời gian (AC-002)
	GetOwnershipHistoryByProductItemID(ctx context.Context, productItemID uuid.UUID) ([]entity.Ownership, error)
}

type OwnershipRepository struct {
	db *gorm.DB
}

func NewOwnershipRepository(db *gorm.DB) IOwnershipRepository {
	return &OwnershipRepository{
		db: db,
	}
}

func (r *OwnershipRepository) CreateOwnership(ctx context.Context, tx *gorm.DB, ownership *entity.Ownership) (*entity.Ownership, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	db = db.WithContext(ctx)
	if err := db.Create(ownership).Error; err != nil {
		return nil, err
	}
	return ownership, nil
}

// GetOwnershipByProductItemID lấy record sở hữu active / mới nhất của product item
func (r *OwnershipRepository) GetOwnershipByProductItemID(ctx context.Context, productItemID uuid.UUID) (*entity.Ownership, error) {
	var ownership entity.Ownership
	if err := r.db.WithContext(ctx).
		Where("product_item_id = ?", productItemID).
		Order("owned_at DESC").
		First(&ownership).Error; err != nil {
		return nil, err
	}
	return &ownership, nil
}

// GetOwnershipHistoryByProductItemID lấy toàn bộ lịch sử đăng ký theo thứ tự thời gian (cũ → mới)
func (r *OwnershipRepository) GetOwnershipHistoryByProductItemID(ctx context.Context, productItemID uuid.UUID) ([]entity.Ownership, error) {
	var history []entity.Ownership
	if err := r.db.WithContext(ctx).
		Where("product_item_id = ?", productItemID).
		Order("owned_at ASC").
		Find(&history).Error; err != nil {
		return nil, err
	}
	return history, nil
}
