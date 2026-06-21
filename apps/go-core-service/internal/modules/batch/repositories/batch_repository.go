package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
)

type BatchRepository interface {
	FindAll(ctx context.Context) ([]*response.BatchListResponse, error)
	FindByCode(ctx context.Context, batchCode string) (*response.BatchDetailResponse, error)
	FindByBatchID(ctx context.Context, batchID uuid.UUID) (*response.BatchDetailResponse, error)
	Create(ctx context.Context, req *request.CreateBatchRequest) (*response.BatchCreateResponse, error)
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

func (rb *batchRepository) FindByBatchID(ctx context.Context, batchID uuid.UUID) (*response.BatchDetailResponse, error) {
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
		WHERE b.id = $1
		  AND b.is_deleted = FALSE
		LIMIT 1`

	row := rb.db.QueryRowContext(ctx, sqlQuery, batchID)

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

// Create sinh batch code tự động theo format [PREFIX]-[YEAR]-[SEQUENCE]
// và dùng PostgreSQL Advisory Lock (transaction-scoped) để đảm bảo
// sequence tăng tuần tự, không bị race condition.
//
// Ví dụ: prefix "APL" + năm 2026 → lock key "APL-2026" → batch code "APL-2026-0013"
// Các prefix khác nhau (SAM-2026, XMI-2026) vẫn chạy song song.
func (rb *batchRepository) Create(ctx context.Context, req *request.CreateBatchRequest) (*response.BatchCreateResponse, error) {
	now := time.Now()
	year := now.Year()
	// lockKey là đơn vị granularity của advisory lock
	lockKey := fmt.Sprintf("%s-%d", req.Prefix, year)

	// Bắt đầu transaction — advisory lock pg_advisory_xact_lock sẽ
	// tự động release khi transaction kết thúc (commit hoặc rollback).
	tx, err := rb.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperror.WrapDBError(err, "batch")
	}
	defer tx.Rollback() //nolint:errcheck

	// Acquire session-level advisory lock trong phạm vi transaction.
	// hashtext() chuyển string → int4 để dùng làm lock key.
	// Nếu lock đang bị giữ bởi transaction khác, câu lệnh này sẽ BLOCK cho đến khi available.
	_, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", lockKey)
	if err != nil {
		return nil, apperror.WrapDBError(err, "batch")
	}

	// Tìm batch_code lớn nhất hiện tại cho prefix+year này.
	// Dùng COALESCE để luôn trả về 1 row (chuỗi rỗng nếu chưa có batch nào).
	pattern := lockKey + "-%"
	var lastCode string
	err = tx.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(batch_code), '') FROM batches WHERE batch_code LIKE $1`,
		pattern,
	).Scan(&lastCode)
	if err != nil {
		return nil, apperror.WrapDBError(err, "batch")
	}

	// Tính sequence tiếp theo.
	nextSeq := 1
	if lastCode != "" {
		// lastCode có dạng "APL-2026-0012" → cắt phần sau lockKey+"-"
		seqStr := strings.TrimPrefix(lastCode, lockKey+"-")
		if seq, parseErr := strconv.Atoi(seqStr); parseErr == nil {
			nextSeq = seq + 1
		}
	}

	// Sinh batch code mới với padding 4 chữ số (APL-2026-0013).
	newBatchCode := fmt.Sprintf("%s-%04d", lockKey, nextSeq)
	newID := uuid.New()

	// INSERT và lấy ngay dữ liệu vừa tạo qua RETURNING.
	insertSQL := `
		INSERT INTO batches (
			id, variant_id, batch_code,
			manufacture_date, expiry_date, imported_at,
			manufacturer_name, supplier_name, origin_country, production_place,
			quantity, status,
			created_at, updated_at, is_deleted
		) VALUES (
			$1,  $2,  $3,
			$4,  $5,  $6,
			$7,  $8,  $9,  $10,
			$11, 'ACTIVE',
			$12, $12, FALSE
		)
		RETURNING id, batch_code, variant_id, quantity, status, created_at`

	row := tx.QueryRowContext(ctx, insertSQL,
		newID,
		req.VariantID,
		newBatchCode,
		req.ManufactureDate,
		req.ExpiryDate,
		req.ImportedAt,
		req.ManufacturerName,
		req.SupplierName,
		req.OriginCountry,
		req.ProductionPlace,
		req.Quantity,
		now,
	)

	var result response.BatchCreateResponse
	if err = row.Scan(
		&result.ID,
		&result.BatchCode,
		&result.VariantID,
		&result.Quantity,
		&result.Status,
		&result.CreatedAt,
	); err != nil {
		log.Println("[batch] insert error:", err)
		return nil, apperror.WrapDBError(err, "batch")
	}

	if err = tx.Commit(); err != nil {
		return nil, apperror.WrapDBError(err, "batch")
	}

	return &result, nil
}
