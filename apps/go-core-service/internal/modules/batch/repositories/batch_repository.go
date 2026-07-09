package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	pkgResponse "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
	"gorm.io/gorm"
)

// BatchRepository định nghĩa toàn bộ contract data-access cho Batch.
// Tất cả implementations đều dùng GORM — không còn raw *sql.DB.
type BatchRepository interface {
	// --- Read ---
	FindAllWithFilter(ctx context.Context, req *request.GetBatchListRequest) (*response.BatchListResponse, error)
	// SearchBatches thực hiện tìm kiếm gần đúng (ILIKE) theo UC-P2-BATCH-03.
	SearchBatches(ctx context.Context, req *request.SearchBatchRequest) (*response.SearchBatchResponse, error)
	FindByCode(ctx context.Context, batchCode string) (*response.BatchDetailResponse, error)
	FindByBatchID(ctx context.Context, batchID uuid.UUID) (*response.BatchDetailResponse, error)
	GetBatchEvents(ctx context.Context, batchID uuid.UUID) ([]response.BatchEventDTO, error)
	// FindByID trả về entity đầy đủ (kể cả đã soft-delete) để service tự kiểm tra is_deleted.
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Batch, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)

	// --- Write ---
	// Create sinh batch code tự động (advisory lock) và insert vào DB.
	Create(ctx context.Context, req *request.CreateBatchRequest, currentUserID uuid.UUID) (*response.BatchCreateResponse, error)
	// UpdateStatus cập nhật duy nhất field status; GORM tự set updated_at.
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	// SoftDelete set is_deleted = true; GORM tự set updated_at.
	SoftDelete(ctx context.Context, id uuid.UUID) error
	// ExportBatch handle export logic with transaction
	ExportBatch(ctx context.Context, batchID uuid.UUID, exportReq *request.ExportBatchRequest, currentUserID uuid.UUID) error

	// --- Constraint checks (dùng trước khi delete) ---
	ExistsProductItems(ctx context.Context, batchID uuid.UUID) (bool, error)
	ExistsEvents(ctx context.Context, batchID uuid.UUID) (bool, error)

	// --- History & Products (UC-P2-BATCH-05, UC-P2-BATCH-06) ---
	// GetBatchHistory trả về lịch sử thay đổi từ bảng audit_logs JOIN users.
	GetBatchHistory(ctx context.Context, batchID uuid.UUID, page, limit int) ([]response.BatchHistoryItemDTO, error)
	// GetBatchProducts trả về danh sách product items có pagination, filter và search.
	GetBatchProducts(ctx context.Context, batchID uuid.UUID, req *request.GetBatchProductsRequest) (*response.GetBatchProductsResponse, error)
}

type batchRepository struct {
	db *gorm.DB
}

// NewBatchRepository tạo BatchRepository backed bởi GORM.
func NewBatchRepository(db *gorm.DB) BatchRepository {
	return &batchRepository{db: db}
}

// ---------------------------------------------------------------------------
// Read methods
// ---------------------------------------------------------------------------

// FindAllWithFilter trả về danh sách batch kèm theo phân trang và filter
func (r *batchRepository) FindAllWithFilter(ctx context.Context, req *request.GetBatchListRequest) (*response.BatchListResponse, error) {
	query := r.db.WithContext(ctx).
		Table("batches b").
		Joins("JOIN product_variants pv ON pv.id = b.variant_id").
		Where("b.is_deleted = false")

	if req.Search != "" {
		search := "%" + req.Search + "%"
		query = query.Where("b.batch_code ILIKE ? OR pv.name ILIKE ? OR b.origin_country ILIKE ?", search, search, search)
	}

	if req.Status != "" && req.Status != "ALL" {
		query = query.Where("b.status = ?", req.Status)
	}

	if req.OriginCountry != "" && req.OriginCountry != "ALL" {
		query = query.Where("b.origin_country = ?", req.OriginCountry)
	}

	// BR-FIL-001: ẩn DRAFT khi user không phải Admin và không filter status cụ thể.
	// ExcludeDraft được service set dựa theo role.
	if req.ExcludeDraft {
		query = query.Where("b.status != ?", "DRAFT")
	}

	var stats response.BatchStatsDTO
	// Calculate stats
	statsQuery := r.db.WithContext(ctx).Table("batches b").Joins("JOIN product_variants pv ON pv.id = b.variant_id").Where("b.is_deleted = false")
	if req.Search != "" {
		search := "%" + req.Search + "%"
		statsQuery = statsQuery.Where("b.batch_code ILIKE ? OR pv.name ILIKE ? OR b.origin_country ILIKE ?", search, search, search)
	}
	if req.OriginCountry != "" && req.OriginCountry != "ALL" {
		statsQuery = statsQuery.Where("b.origin_country = ?", req.OriginCountry)
	}
	// Áp dụng ExcludeDraft vào stats query để thống kê nhất quán với kết quả trả về.
	if req.ExcludeDraft {
		statsQuery = statsQuery.Where("b.status != ?", "DRAFT")
	}

	// We must use case when to calculate stats
	statsQuery.Select(`
		COUNT(*) as total,
		SUM(CASE WHEN b.status = 'ACTIVE' THEN 1 ELSE 0 END) as active,
		SUM(CASE WHEN b.status = 'EXPIRED' THEN 1 ELSE 0 END) as expired,
		SUM(CASE WHEN b.status IN ('RECALLED', 'BLOCKED') THEN 1 ELSE 0 END) as recalled_blocked
	`).Scan(&stats)

	var totalItems int64
	query.Count(&totalItems)

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	var items []response.BatchListItemDTO
	err := query.
		Select(`
			b.id, b.batch_code, b.variant_id, pv.name as variant_name, 
			b.quantity, b.manufacture_date, b.expiry_date, 
			b.origin_country, b.status
		`).
		Order("b.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&items).Error

	if err != nil {
		return nil, apperror.WrapDBError(err, "batch")
	}

	totalPages := int((totalItems + int64(limit) - 1) / int64(limit))
	if totalPages == 0 && totalItems > 0 {
		totalPages = 1
	}

	meta := pkgResponse.PaginationMeta{
		CurrentPage: page,
		PageSize:    limit,
		TotalItems:  int(totalItems),
		TotalPages:  totalPages,
	}

	return &response.BatchListResponse{
		Items: items,
		Meta:  meta,
		Stats: stats,
	}, nil
}

// ---------------------------------------------------------------------------
// Search (UC-P2-BATCH-03)
// ---------------------------------------------------------------------------

// searchSortFields là whitelist để tránh SQL Injection khi build ORDER BY động.
var searchSortFields = map[string]string{
	"createdAt": "b.created_at",
	"batchCode": "b.batch_code",
}

// SearchBatches thực hiện tìm kiếm gần đúng (ILIKE) theo keyword trên các trường
// batch_code, manufacturer_name và product_variant.name.
//
// Các điều kiện filter chỉ được apply khi client truyền vào (không sinh WHERE dư thừa).
// sortBy và sortOrder được validate qua whitelist để đảm bảo an toàn.
func (r *batchRepository) SearchBatches(ctx context.Context, req *request.SearchBatchRequest) (*response.SearchBatchResponse, error) {
	query := r.db.WithContext(ctx).
		Table("batches b").
		Joins("JOIN product_variants pv ON pv.id = b.variant_id").
		Joins("JOIN products p ON p.id = pv.product_id").
		Where("b.is_deleted = false")

	// Apply keyword filter (ILIKE — case-insensitive, BR-SEA-002)
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		query = query.Where(
			"b.batch_code ILIKE ? OR pv.name ILIKE ? OR b.manufacturer_name ILIKE ?",
			kw, kw, kw,
		)
	}

	// Đếm tổng số bản ghi thỏa điều kiện (không limit/offset)
	var totalRecords int64
	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, apperror.WrapDBError(err, "batch")
	}

	// Tính pagination
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	totalPages := int((totalRecords + int64(pageSize) - 1) / int64(pageSize))
	if totalPages == 0 && totalRecords > 0 {
		totalPages = 1
	}

	// Build ORDER BY — validate sortBy qua whitelist để tránh SQL Injection
	sortCol, ok := searchSortFields[req.SortBy]
	if !ok {
		sortCol = "b.created_at"
	}
	sortDir := "DESC"
	if strings.EqualFold(req.SortOrder, "ASC") {
		sortDir = "ASC"
	}
	orderClause := sortCol + " " + sortDir

	// Scan vào struct tạm để tránh conflict tên cột
	type searchRow struct {
		ID              uuid.UUID
		BatchCode       string
		ProductName     string
		ManufactureDate *time.Time
		Quantity        int
		Status          string
		CreatedAt       time.Time
	}

	var rows []searchRow
	err := query.
		Select(`
			b.id,
			b.batch_code,
			p.name  AS product_name,
			b.manufacture_date,
			b.quantity,
			b.status,
			b.created_at
		`).
		Order(orderClause).
		Limit(pageSize).
		Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, apperror.WrapDBError(err, "batch")
	}

	// Map sang DTO
	items := make([]response.SearchBatchItemDTO, 0, len(rows))
	for _, row := range rows {
		items = append(items, response.SearchBatchItemDTO{
			BatchID:           row.ID,
			BatchCode:         row.BatchCode,
			ProductName:       row.ProductName,
			ManufacturingDate: row.ManufactureDate,
			Quantity:          row.Quantity,
			Status:            row.Status,
			CreatedAt:         row.CreatedAt,
		})
	}

	return &response.SearchBatchResponse{
		Items: items,
		Pagination: response.SearchBatchPaginationDTO{
			CurrentPage:  page,
			PageSize:     pageSize,
			TotalRecords: int(totalRecords),
			TotalPages:   totalPages,
		},
	}, nil
}

// FindByCode trả về chi tiết batch theo batch_code, kèm variant và product.
func (r *batchRepository) FindByCode(ctx context.Context, batchCode string) (*response.BatchDetailResponse, error) {
	return r.findDetail(ctx, "b.batch_code = ?", batchCode)
}

// FindByBatchID trả về chi tiết batch theo UUID, kèm variant và product.
func (r *batchRepository) FindByBatchID(ctx context.Context, batchID uuid.UUID) (*response.BatchDetailResponse, error) {
	return r.findDetail(ctx, "b.id = ?", batchID)
}

// findDetail là helper dùng chung cho FindByCode và FindByBatchID.
func (r *batchRepository) findDetail(ctx context.Context, condition string, value any) (*response.BatchDetailResponse, error) {
	type row struct {
		ID               uuid.UUID
		BatchCode        string
		ManufactureDate  *time.Time
		ExpiryDate       *time.Time
		ImportedAt       *time.Time
		ManufacturerName *string
		SupplierName     *string
		OriginCountry    *string
		ProductionPlace  *string
		Quantity         int
		Status           string
		CreatedAt        time.Time
		UpdatedAt        time.Time
		VariantID        uuid.UUID
		VariantSKU       string
		VariantName      string
		VariantBarcode   *string
		ProductID        uuid.UUID
		ProductName      string
	}

	var r2 row
	err := r.db.WithContext(ctx).
		Table("batches b").
		Select(`
			b.id, b.batch_code,
			b.manufacture_date, b.expiry_date, b.imported_at,
			b.manufacturer_name, b.supplier_name, b.origin_country, b.production_place,
			b.quantity, b.status, b.created_at, b.updated_at,
			pv.id   AS variant_id,
			pv.sku  AS variant_sku,
			pv.name AS variant_name,
			pv.barcode AS variant_barcode,
			p.id    AS product_id,
			p.name  AS product_name
		`).
		Joins("JOIN product_variants pv ON pv.id = b.variant_id").
		Joins("JOIN products p ON p.id = pv.product_id").
		Where(condition, value).
		Where("b.is_deleted = false").
		Limit(1).
		Scan(&r2).Error
	if err != nil {
		return nil, apperror.WrapDBError(err, "batch")
	}
	if r2.ID == (uuid.UUID{}) {
		return nil, apperror.WrapDBError(gorm.ErrRecordNotFound, "batch")
	}

	return &response.BatchDetailResponse{
		ID:               r2.ID,
		BatchCode:        r2.BatchCode,
		ManufactureDate:  r2.ManufactureDate,
		ExpiryDate:       r2.ExpiryDate,
		ImportedAt:       r2.ImportedAt,
		ManufacturerName: r2.ManufacturerName,
		SupplierName:     r2.SupplierName,
		OriginCountry:    r2.OriginCountry,
		ProductionPlace:  r2.ProductionPlace,
		Quantity:         r2.Quantity,
		Status:           r2.Status,
		CreatedAt:        r2.CreatedAt,
		UpdatedAt:        r2.UpdatedAt,
		Variant: response.BatchDetailVariantResponse{
			VariantID: r2.VariantID,
			SKU:       r2.VariantSKU,
			Name:      r2.VariantName,
			Barcode:   r2.VariantBarcode,
		},
		Product: response.BatchDetailProductResponse{
			ProductID:   r2.ProductID,
			ProductName: r2.ProductName,
		},
	}, nil
}

// FindByID trả về entity Batch (kể cả đã soft-delete) để service kiểm tra is_deleted.
func (r *batchRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Batch, error) {
	var batch entities.Batch
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&batch).Error
	if err != nil {
		return nil, apperror.WrapDBError(err, "batch")
	}
	return &batch, nil
}

// ExistsByID kiểm tra batch tồn tại và chưa bị xóa mềm.
func (r *batchRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Batch{}).
		Where("id = ? AND is_deleted = false", id).
		Count(&count).Error
	if err != nil {
		return false, apperror.WrapDBError(err, "batch")
	}
	return count > 0, nil
}

func (r *batchRepository) GetBatchEvents(ctx context.Context, batchID uuid.UUID) ([]response.BatchEventDTO, error) {
	var events []response.BatchEventDTO

	err := r.db.WithContext(ctx).
		Table("events e").
		Select("e.event_type as event_name, e.description as detail, e.created_at").
		Joins("JOIN product_items p ON e.product_item_id = p.id").
		Where("p.batch_id = ? AND e.is_deleted = false", batchID).
		Order("e.created_at DESC").
		Scan(&events).Error

	if err != nil {
		return nil, apperror.WrapDBError(err, "events")
	}

	if events == nil {
		events = []response.BatchEventDTO{}
	}

	return events, nil
}

// ---------------------------------------------------------------------------
// Write methods
// ---------------------------------------------------------------------------

// Create sinh batch code tự động theo format [PREFIX]-[YEAR]-[SEQUENCE]
// và dùng PostgreSQL Advisory Lock (transaction-scoped) để đảm bảo
// sequence tăng tuần tự, không bị race condition.
//
// Ví dụ: prefix "APL" + năm 2026 → lock key "APL-2026" → batch code "APL-2026-0013"
// Các prefix khác nhau (SAM-2026, XMI-2026) vẫn chạy song song vì lock key khác nhau.
func (r *batchRepository) Create(ctx context.Context, req *request.CreateBatchRequest, currentUserID uuid.UUID) (*response.BatchCreateResponse, error) {
	now := time.Now()
	year := now.Year()
	lockKey := fmt.Sprintf("%s-%d", req.Prefix, year)

	var result response.BatchCreateResponse

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Acquire PostgreSQL advisory lock (transaction-scoped).
		//    Tự động release khi transaction kết thúc (commit hoặc rollback).
		//    hashtext() chuyển string → int4 để dùng làm lock key.
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", lockKey).Error; err != nil {
			return apperror.WrapDBError(err, "batch")
		}

		// 2. Tìm batch_code lớn nhất cho prefix+year này.
		pattern := lockKey + "-%"
		var lastCode string
		if err := tx.Raw(
			"SELECT COALESCE(MAX(batch_code), '') FROM batches WHERE batch_code LIKE ?",
			pattern,
		).Scan(&lastCode).Error; err != nil {
			return apperror.WrapDBError(err, "batch")
		}

		// 3. Tính sequence tiếp theo.
		nextSeq := 1
		if lastCode != "" {
			seqStr := strings.TrimPrefix(lastCode, lockKey+"-")
			if seq, parseErr := strconv.Atoi(seqStr); parseErr == nil {
				nextSeq = seq + 1
			}
		}

		// 4. Sinh batch code mới với padding 4 chữ số.
		newBatchCode := fmt.Sprintf("%s-%04d", lockKey, nextSeq)
		newID := uuid.New()

		batch := entities.Batch{
			ID:               newID,
			VariantID:        req.VariantID,
			BatchCode:        newBatchCode,
			ManufactureDate:  req.ManufactureDate,
			ExpiryDate:       req.ExpiryDate,
			ImportedAt:       req.ImportedAt,
			ManufacturerName: derefString(req.ManufacturerName),
			SupplierName:     derefString(req.SupplierName),
			OriginCountry:    derefString(req.OriginCountry),
			ProductionPlace:  derefString(req.ProductionPlace),
			Quantity:         req.Quantity,
			Status:           "ACTIVE",
			CreatedBy:        &currentUserID,
			CreatedAt:        now,
		}

		// 5. INSERT via GORM.
		if err := tx.Create(&batch).Error; err != nil {
			return apperror.WrapDBError(err, "batch")
		}

		result = response.BatchCreateResponse{
			ID:        batch.ID,
			BatchCode: batch.BatchCode,
			VariantID: batch.VariantID,
			Quantity:  batch.Quantity,
			Status:    batch.Status,
			CreatedAt: batch.CreatedAt,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateStatus cập nhật field status; GORM tự set updated_at (autoUpdateTime).
// Repository không validate enum — đó là trách nhiệm của Service.
func (r *batchRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	result := r.db.WithContext(ctx).
		Model(&entities.Batch{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return apperror.WrapDBError(result.Error, "batch")
	}
	return nil
}

// SoftDelete set is_deleted = true; GORM tự set updated_at.
func (r *batchRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&entities.Batch{}).
		Where("id = ?", id).
		Update("is_deleted", true)
	if result.Error != nil {
		return apperror.WrapDBError(result.Error, "batch")
	}
	return nil
}

func (r *batchRepository) ExportBatch(ctx context.Context, batchID uuid.UUID, exportReq *request.ExportBatchRequest, currentUserID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var batch entities.Batch
		if err := tx.Where("id = ? AND is_deleted = false", batchID).First(&batch).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return apperror.NewNotFound("batch not found")
			}
			return apperror.WrapDBError(err, "batch")
		}

		if batch.Quantity < exportReq.Quantity {
			return apperror.NewBadRequest("insufficient quantity to export")
		}

		batch.Quantity -= exportReq.Quantity
		if err := tx.Save(&batch).Error; err != nil {
			return apperror.WrapDBError(err, "batch")
		}

		// Because events table require product_item_id which may not be appropriate for a pure batch export
		// (without selecting specific items), and there's no batch_events table.
		// We insert an audit_log record for the export action.

		// Wait, user says "Create history/event. Create audit log."
		// Since we don't have batch_id in events and user said status is just reference and event might not be strictly needed for batch,
		// we will just insert an audit_log to represent the history.
		// "audit_logs" schema from migration: id, action, entity_type, entity_id, user_id, old_values, new_values, ip_address, created_at
		auditLog := map[string]interface{}{
			"id":          uuid.New(),
			"action":      "EXPORT_BATCH",
			"entity_type": "BATCH",
			"entity_id":   batch.ID.String(),
			"user_id":     currentUserID.String(),
			"new_values":  fmt.Sprintf(`{"exported_quantity": %d, "destination": "%s", "operator": "%s", "notes": "%s"}`, exportReq.Quantity, exportReq.DestinationLocation, exportReq.OperatorName, exportReq.Notes),
			"created_at":  time.Now(),
		}

		if err := tx.Table("audit_logs").Create(auditLog).Error; err != nil {
			return apperror.WrapDBError(err, "audit_log")
		}

		return nil
	})
}

// ---------------------------------------------------------------------------
// History & Products (UC-P2-BATCH-06, UC-P2-BATCH-05)
// ---------------------------------------------------------------------------

// GetBatchHistory trả về danh sách lịch sử thay đổi của một Batch từ bảng audit_logs.
// JOIN LEFT bảng users để lấy full_name và role của người thực hiện.
// old_data và new_data được trả thồ ngà rư để service parse sang diff-view.
func (r *batchRepository) GetBatchHistory(ctx context.Context, batchID uuid.UUID, page, limit int) ([]response.BatchHistoryItemDTO, error) {
	type historyRow struct {
		LogID     uuid.UUID
		Action    string
		OldData   []byte
		NewData   []byte
		CreatedAt time.Time
		// User fields (nullable — system actions có thể không có user)
		UserID   *uuid.UUID
		FullName *string
		Role     *string
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 15
	}
	offset := (page - 1) * limit

	var rows []historyRow
	err := r.db.WithContext(ctx).
		Table("audit_logs al").
		Select(`
			al.id  AS log_id,
			al.action,
			al.old_data,
			al.new_data,
			al.created_at,
			u.id        AS user_id,
			u.full_name AS full_name,
			u.role      AS role
		`).
		Joins("LEFT JOIN users u ON u.id = al.user_id").
		Where("al.entity = ? AND al.entity_id = ?", "Batch", batchID).
		Order("al.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, apperror.WrapDBError(err, "audit_logs")
	}

	items := make([]response.BatchHistoryItemDTO, 0, len(rows))
	for _, row := range rows {
		// Parse old_data và new_data từ JSONB sang map[string]any
		var oldMap, newMap map[string]any
		if len(row.OldData) > 0 {
			_ = json.Unmarshal(row.OldData, &oldMap)
		}
		if len(row.NewData) > 0 {
			_ = json.Unmarshal(row.NewData, &newMap)
		}

		// Build changedFields: xây dựng diff-view bằng cách so sánh key giữa old và new.
		// Chỉ show key xuất hiện trong newMap (trường thực sự đã được cập nhật).
		changedFields := make(map[string]response.FieldDiffDTO)
		for key, newVal := range newMap {
			oldVal := oldMap[key]
			changedFields[key] = response.FieldDiffDTO{
				Old: oldVal,
				New: newVal,
			}
		}

		// Map performer (nil nếu system action)
		var actor *response.BatchHistoryActorDTO
		if row.UserID != nil && row.FullName != nil {
			role := ""
			if row.Role != nil {
				role = *row.Role
			}
			actor = &response.BatchHistoryActorDTO{
				UserID:   *row.UserID,
				FullName: *row.FullName,
				Role:     role,
			}
		}

		items = append(items, response.BatchHistoryItemDTO{
			LogID:         row.LogID,
			Action:        row.Action,
			ChangedFields: changedFields,
			PerformedBy:   actor,
			IPAddress:     "",
			CreatedAt:     row.CreatedAt,
		})
	}

	return items, nil
}

// validProductItemStatuses là whitelist trạng thái hợp lệ của product_item để
// tránh SQL Injection khi filter status trong GetBatchProducts.
var validProductItemStatuses = map[string]struct{}{
	"AVAILABLE":  {},
	"IN_TRANSIT": {},
	"SOLD":       {},
	"RECALLED":   {},
}

// GetBatchProducts trả về danh sách sản phẩm đơn lả trong lô với pagination.
// LEFT JOIN locations để lấy thông tin vị trí lưu kho hiện tại.
// Sắp xếp theo serial_number ASC theo spec UC-P2-BATCH-05.
func (r *batchRepository) GetBatchProducts(ctx context.Context, batchID uuid.UUID, req *request.GetBatchProductsRequest) (*response.GetBatchProductsResponse, error) {
	query := r.db.WithContext(ctx).
		Table("product_items pi").
		Joins("LEFT JOIN locations l ON l.id = pi.current_location_id").
		Where("pi.batch_id = ? AND pi.is_deleted = false", batchID)

	// Filter status (chỉ apply khi client truyền vào và hợp lệ)
	if req.Status != "" {
		if _, ok := validProductItemStatuses[req.Status]; ok {
			query = query.Where("pi.status = ?", req.Status)
		}
	}

	// Search keyword ILIKE trên item_code và serial_number (AF-001)
	if req.Keyword != "" {
		kw := "%" + req.Keyword + "%"
		query = query.Where("pi.item_code ILIKE ? OR pi.serial_number ILIKE ?", kw, kw)
	}

	var totalRecords int64
	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, apperror.WrapDBError(err, "product_items")
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit
	totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))
	if totalPages == 0 && totalRecords > 0 {
		totalPages = 1
	}

	// Struct tạm để Scan — tránh conflict tên cột giữa các bảng.
	type productRow struct {
		ItemID         uuid.UUID
		ItemCode       string
		SerialNumber   string
		Status         string
		CreatedAt      time.Time
		LocationID     *uuid.UUID
		LocationName   *string
		LocationType   *string
	}

	var rows []productRow
	err := query.
		Select(`
			pi.id             AS item_id,
			pi.item_code,
			pi.serial_number,
			pi.status,
			pi.created_at,
			l.id              AS location_id,
			l.name            AS location_name,
			l.type            AS location_type
		`).
		Order("pi.serial_number ASC").
		Limit(limit).
		Offset(offset).
		Scan(&rows).Error
	if err != nil {
		return nil, apperror.WrapDBError(err, "product_items")
	}

	// Map sang DTO — xử lý nullable location (LEFT JOIN)
	items := make([]response.BatchProductItemDTO, 0, len(rows))
	for _, row := range rows {
		var location *response.BatchProductLocationDTO
		if row.LocationID != nil {
			location = &response.BatchProductLocationDTO{
				LocationID: *row.LocationID,
				Name:       derefString(row.LocationName),
				Type:       derefString(row.LocationType),
			}
		}
		items = append(items, response.BatchProductItemDTO{
			ItemID:          row.ItemID,
			ItemCode:        row.ItemCode,
			SerialNumber:    row.SerialNumber,
			Status:          row.Status,
			CurrentLocation: location,
			CreatedAt:       row.CreatedAt,
		})
	}

	return &response.GetBatchProductsResponse{
		Items: items,
		Pagination: response.BatchProductPaginationDTO{
			CurrentPage:  page,
			PageSize:     limit,
			TotalRecords: int(totalRecords),
			TotalPages:   totalPages,
		},
	}, nil
}

// ---------------------------------------------------------------------------
// Constraint checks
// ---------------------------------------------------------------------------

// ExistsProductItems kiểm tra có ít nhất 1 product item liên kết (chưa bị xóa).
func (r *batchRepository) ExistsProductItems(ctx context.Context, batchID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("product_items").
		Where("batch_id = ? AND is_deleted = false", batchID).
		Count(&count).Error
	if err != nil {
		return false, apperror.WrapDBError(err, "product_items")
	}
	return count > 0, nil
}

// ExistsEvents kiểm tra có event nào liên kết qua product_items với batch này.
func (r *batchRepository) ExistsEvents(ctx context.Context, batchID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("events").
		Joins("JOIN product_items ON product_items.id = events.product_item_id").
		Where("product_items.batch_id = ? AND events.is_deleted = false", batchID).
		Count(&count).Error
	if err != nil {
		return false, apperror.WrapDBError(err, "events")
	}
	return count > 0, nil
}

// ---------------------------------------------------------------------------
// Private helpers
// ---------------------------------------------------------------------------

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
