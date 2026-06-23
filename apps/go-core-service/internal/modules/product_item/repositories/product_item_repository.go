package repositories

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"gorm.io/gorm"
)

type ProductItemRepository interface {
	FindByBatchID(ctx context.Context, batchID uuid.UUID) ([]*entities.ProductItem, error)
	Create(ctx context.Context, pi *entities.ProductItem) (*response.ProductItemCreateResponse, error)
}

type productItemRepository struct {
	db    *gorm.DB
	sqlDB *sql.DB
}

func NewProductItemRepository(db *gorm.DB, sqlDB *sql.DB) ProductItemRepository {
	return &productItemRepository{
		db:    db,
		sqlDB: sqlDB,
	}
}

func (rp *productItemRepository) FindByBatchID(ctx context.Context, batchID uuid.UUID) ([]*entities.ProductItem, error) {
	productItems, err := gorm.G[*entities.ProductItem](rp.db).Where("batch_id = ?", batchID).Find(ctx)
	if err != nil {
		return nil, apperror.WrapDBError(err, "product_items")
	}

	return productItems, nil
}

// Create tạo một product item mới với 3 trường định danh được tự động sinh:
//   - item_code:          PTA-{YYMM}-{8 ký tự HEX viết hoa}
//   - serial_number:      SN{14 chữ số ngẫu nhiên}
//   - verification_token: MD5 hex 32 ký tự thường
//
// Dùng raw SQL với RETURNING để tránh N+1 query và lấy ngay dữ liệu đã lưu.
func (rp *productItemRepository) Create(ctx context.Context, pi *entities.ProductItem) (*response.ProductItemCreateResponse, error) {

	insertSQL := `
		INSERT INTO product_items (
			id, variant_id, batch_id,
			item_code, serial_number, verification_token,
			status,
			created_at, updated_at, is_deleted
		) VALUES (
			$1, $2, $3,
			$4, $5, $6,
			'ACTIVE',
			$7, $7, FALSE
		)
		RETURNING id, variant_id, batch_id, item_code, serial_number, verification_token, status, created_at`

	row := rp.sqlDB.QueryRowContext(ctx, insertSQL,
		pi.ID,
		pi.VariantID,
		pi.BatchID,
		pi.ItemCode,
		pi.SerialNumber,
		pi.VerificationToken,
		time.Now(),
	)

	var result response.ProductItemCreateResponse
	err := row.Scan(
		&result.ID,
		&result.VariantID,
		&result.BatchID,
		&result.ItemCode,
		&result.SerialNumber,
		&result.VerificationToken,
		&result.Status,
		&result.CreatedAt,
	)
	if err != nil {
		log.Println("[product_item] insert error:", err)
		return nil, apperror.WrapDBError(err, "product_items")
	}

	return &result, nil
}
