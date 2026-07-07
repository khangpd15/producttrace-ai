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

	// CRUD Extensions
	GetOwnershipByID(ctx context.Context, id uuid.UUID) (*entity.Ownership, error)
	UpdateOwnershipStatusAndEndedAt(ctx context.Context, tx *gorm.DB, id uuid.UUID, status string) error
	
	// Transaction Support
	BeginTx(ctx context.Context) *gorm.DB
	
	// Search & List
	SearchOwnerships(ctx context.Context, filter SearchFilter) ([]entity.Ownership, int64, error)

	// TransferOwnershipTx marks old as TRANSFERRED and inserts new — atomically
	TransferOwnershipTx(ctx context.Context, oldID uuid.UUID, newOwnership *entity.Ownership) error
}

type SearchFilter struct {
	ProductItemIDs  []uuid.UUID
	OwnerIDs        []uuid.UUID
	OwnershipStatus string
	Page            int
	Limit           int
	Role            string
	CurrentUserID   uuid.UUID
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

// BeginTx khởi tạo một transaction cho các thao tác cross-table
func (r *OwnershipRepository) BeginTx(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Begin()
}

// GetOwnershipByID lấy Ownership record bằng primary ID
func (r *OwnershipRepository) GetOwnershipByID(ctx context.Context, id uuid.UUID) (*entity.Ownership, error) {
	var ownership entity.Ownership
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&ownership).Error; err != nil {
		return nil, err
	}
	return &ownership, nil
}

// UpdateOwnershipStatusAndEndedAt update status và set ended_at = Now (dùng cho Transfer, Delete)
func (r *OwnershipRepository) UpdateOwnershipStatusAndEndedAt(ctx context.Context, tx *gorm.DB, id uuid.UUID, status string) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	
	// gorm hooks or updates multiple fields securely via map
	return db.WithContext(ctx).Model(&entity.Ownership{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":   status,
			"ended_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}

// SearchOwnerships API danh sách + bộ lọc tìm kiếm
func (r *OwnershipRepository) SearchOwnerships(ctx context.Context, filter SearchFilter) ([]entity.Ownership, int64, error) {
	db := r.db.WithContext(ctx).Model(&entity.Ownership{})

	// Filter by Role & Permission
	if filter.Role == "CUSTOMER" {
		db = db.Where("owner_id = ?", filter.CurrentUserID)
	}

	// Dynamic Filters
	if len(filter.ProductItemIDs) > 0 {
		db = db.Where("product_item_id IN ?", filter.ProductItemIDs)
	}
	if len(filter.OwnerIDs) > 0 {
		db = db.Where("owner_id IN ?", filter.OwnerIDs)
	}
	if filter.OwnershipStatus != "" {
		db = db.Where("status = ?", filter.OwnershipStatus)
	}

	// Count Total
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var results []entity.Ownership
	if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

// TransferOwnershipTx atomically: marks old record as TRANSFERRED and inserts the new owner record
func (r *OwnershipRepository) TransferOwnershipTx(ctx context.Context, oldID uuid.UUID, newOwnership *entity.Ownership) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: expire old record
		if err := tx.Model(&entity.Ownership{}).
			Where("id = ?", oldID).
			Updates(map[string]interface{}{
				"status":   string(entity.OwnershipStatusTransferred),
				"ended_at": gorm.Expr("CURRENT_TIMESTAMP"),
			}).Error; err != nil {
			return err
		}
		// Step 2: create new owner record
		return tx.Create(newOwnership).Error
	})
}
