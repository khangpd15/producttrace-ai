package repositories

import (
	"context"
	"database/sql"
	"log"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
)

type BatchRepository interface {
	FindAll(ctx context.Context) ([]*response.BatchListResponse, error)
	FindByCode(ctx context.Context, batchCode string) (*response.BatchDetailResponse, error)
}

type batchRepository struct {
	db *sql.DB
}

func NewBatchRepository(db *sql.DB) BatchRepository {
	return &batchRepository{
		db: db,
	}
}

func (rb *batchRepository) FindAll(ctx context.Context) ([]*response.BatchListResponse, error) {
	batches := make([]*response.BatchListResponse, 0)
	sqlQuery := `select b.batch_code, pv.product_id, b.status, b.created_at,  b.expiry_date from batches b 
				join product_variants pv on pv.id = b.variant_id
				where b.is_deleted = FALSE`
	rows, err := rb.db.QueryContext(ctx, sqlQuery)

	if err != nil {
		return nil, apperror.WrapDBError(err, "batch")
	}

	defer rows.Close()

	for rows.Next() {
		var batch response.BatchListResponse
		err := rows.Scan(&batch.BatchCode, &batch.ProductID, &batch.Status, &batch.CreatedAt, &batch.ExpiryDate)
		if err != nil {
			return nil, apperror.WrapDBError(err, "batch")
		}
		batches = append(batches, &batch)
	}

	return batches, rows.Err()
}

func (rb *batchRepository) FindByCode(ctx context.Context, batchCode string) (*response.BatchDetailResponse, error) {
	sqlQuery := `
		SELECT
			b.id,
			b.batch_code,
			b.manufacture_date,
			b.expiry_date,
			b.imported_at,
			b.manufacturer_name,
			b.supplier_name,
			b.origin_country,
			b.production_place,
			b.quantity,
			b.status,
			b.created_at,
			b.updated_at,
			pv.id     AS variant_id,
			pv.sku    AS variant_sku,
			pv.name   AS variant_name,
			pv.barcode AS variant_barcode,
			p.id      AS product_id,
			p.name    AS product_name
		FROM batches b
			JOIN product_variants pv ON pv.id = b.variant_id
			JOIN products p ON p.id = pv.product_id
		WHERE b.batch_code = $1
		  AND b.is_deleted = FALSE
		LIMIT 1`

	row := rb.db.QueryRowContext(ctx, sqlQuery, batchCode)

	var detail response.BatchDetailResponse
	err := row.Scan(
		&detail.ID,
		&detail.BatchCode,
		&detail.ManufactureDate,
		&detail.ExpiryDate,
		&detail.ImportedAt,
		&detail.ManufacturerName,
		&detail.SupplierName,
		&detail.OriginCountry,
		&detail.ProductionPlace,
		&detail.Quantity,
		&detail.Status,
		&detail.CreatedAt,
		&detail.UpdatedAt,
		&detail.Variant.VariantID,
		&detail.Variant.SKU,
		&detail.Variant.Name,
		&detail.Variant.Barcode,
		&detail.Product.ProductID,
		&detail.Product.ProductName,
	)
	if err != nil {
		log.Println(err)
		return nil, apperror.WrapDBError(err, "batch")
	}

	return &detail, nil
}
