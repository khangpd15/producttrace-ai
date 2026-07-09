package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"gorm.io/gorm"
)

type ProductItemRepository interface {
	FindByBatchID(ctx context.Context, batchID uuid.UUID) ([]*entities.ProductItem, error)
	Create(ctx context.Context, items []entities.ProductItem) error
}

type productItemRepository struct {
	db *gorm.DB
}

func NewProductItemRepository(db *gorm.DB) ProductItemRepository {
	return &productItemRepository{
		db: db,
	}
}

func (rp *productItemRepository) FindByBatchID(ctx context.Context, batchID uuid.UUID) ([]*entities.ProductItem, error) {
	productItems, err := gorm.G[*entities.ProductItem](rp.db).Where("batch_id = ?", batchID).Find(ctx)
	if err != nil {
		return nil, apperror.WrapDBError(err, "product_items")
	}

	return productItems, nil
}

func (rp *productItemRepository) Create(ctx context.Context, items []entities.ProductItem) error {
	return rp.db.WithContext(ctx).Create(&items).Error
}
