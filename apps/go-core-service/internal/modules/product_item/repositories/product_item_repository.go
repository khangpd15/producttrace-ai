package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	pkgResponse "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
	"gorm.io/gorm"
)

type ProductItemRepository interface {
	FindByBatchID(ctx context.Context, batchID uuid.UUID) ([]*entities.ProductItem, error)
	FindAllWithFilter(ctx context.Context, req *request.GetProductItemListRequest) (*response.ProductItemListResponse, error)
	FindByItemCodeWithEvents(ctx context.Context, itemCode string) (*response.ProductItemDetailDTO, error)
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

func (rp *productItemRepository) FindAllWithFilter(ctx context.Context, req *request.GetProductItemListRequest) (*response.ProductItemListResponse, error) {
	var items []response.ProductItemListItemDTO
	var total int64

	query := rp.db.WithContext(ctx).
		Table("product_items pi").
		Select(`
			pi.id, pi.item_code, 
			p.name as product_name, pv.name as variant_name, 
			b.batch_code, pi.status, pi.created_at
		`).
		Joins("JOIN batches b ON pi.batch_id = b.id").
		Joins("JOIN product_variants pv ON b.variant_id = pv.id").
		Joins("JOIN products p ON pv.product_id = p.id").
		Where("pi.is_deleted = false")

	if req.BatchID != nil {
		query = query.Where("pi.batch_id = ?", req.BatchID)
	}
	if req.Status != nil {
		query = query.Where("pi.status = ?", req.Status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, apperror.WrapDBError(err, "product_items")
	}

	offset := (req.Page - 1) * req.Limit
	err = query.
		Order("pi.created_at DESC").
		Offset(offset).
		Limit(req.Limit).
		Scan(&items).Error

	if err != nil {
		return nil, apperror.WrapDBError(err, "product_items")
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))
	if totalPages == 0 && total > 0 {
		totalPages = 1
	}

	return &response.ProductItemListResponse{
		Items: items,
		Meta: pkgResponse.PaginationMeta{
			CurrentPage: req.Page,
			PageSize:    req.Limit,
			TotalItems:  int(total),
			TotalPages:  totalPages,
		},
	}, nil
}

func (rp *productItemRepository) FindByItemCodeWithEvents(ctx context.Context, itemCode string) (*response.ProductItemDetailDTO, error) {
	var detail response.ProductItemDetailDTO

	err := rp.db.WithContext(ctx).
		Table("product_items pi").
		Select(`
			pi.id, pi.item_code, 
			p.name as product_name, pv.name as variant_name, 
			b.batch_code, pi.status, pi.verification_token, pi.created_at
		`).
		Joins("JOIN batches b ON pi.batch_id = b.id").
		Joins("JOIN product_variants pv ON b.variant_id = pv.id").
		Joins("JOIN products p ON pv.product_id = p.id").
		Where("pi.item_code = ? AND pi.is_deleted = false", itemCode).
		First(&detail).Error

	if err != nil {
		return nil, apperror.WrapDBError(err, "product_items")
	}

	var events []response.ProductItemEventDTO
	err = rp.db.WithContext(ctx).
		Table("events e").
		Select("e.event_type as event_name, e.description as detail, e.created_at").
		Where("e.product_item_id = ? AND e.is_deleted = false", detail.ID).
		Order("e.created_at ASC").
		Scan(&events).Error

	if err != nil {
		return nil, apperror.WrapDBError(err, "events")
	}

	if events == nil {
		events = []response.ProductItemEventDTO{}
	}
	detail.Events = events

	return &detail, nil
}

func (rp *productItemRepository) Create(ctx context.Context, items []entities.ProductItem) error {
	return rp.db.WithContext(ctx).Create(&items).Error
}
