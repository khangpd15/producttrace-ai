package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/xuri/excelize/v2"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/publisher"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/events/types"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/trace/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/trace/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/trace/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/audit_log"
)

type TraceService interface {
	SearchTimeline(ctx context.Context, req *request.TraceSearchRequest, userRole string, clientIP string, userAgent string, userID *uuid.UUID) (*response.TraceSearchResponse, error)
	ExportPDF(ctx context.Context, req *request.PDFExportRequest, userID *uuid.UUID) (*response.ExportJobResponse, []byte, error)
	ExportExcel(ctx context.Context, req *request.ExcelExportRequest, userID *uuid.UUID) (*response.ExportJobResponse, []byte, error)
}

type ItemWithEvents struct {
	Item   *repositories.ProductItemDetail
	Events []repositories.TimelineEvent
}

type traceService struct {
	repo         repositories.TraceRepository
	redisClient  *redis.Client
	pub          *publisher.Publisher
	auditService audit_log.AuditLogService
	baseURL      string
}

func NewTraceService(repo repositories.TraceRepository, redisClient *redis.Client, pub *publisher.Publisher, auditService audit_log.AuditLogService, baseURL string) TraceService {
	return &traceService{
		repo:         repo,
		redisClient:  redisClient,
		pub:          pub,
		auditService: auditService,
		baseURL:      baseURL,
	}
}

// Map valid event types
var validEventTypes = map[string]struct{}{
	"PRODUCED":          {},
	"PACKED":            {},
	"WAREHOUSE_IN":      {},
	"WAREHOUSE_OUT":     {},
	"TRANSPORTED":       {},
	"SALE":              {},
	"REGISTERED":        {},
	"WARRANTY_ACTIVE":   {},
	"WARRANTY_CLAIM":    {},
	"WARRANTY_RESOLVED": {},
	"RETURNED":          {},
	"RECALL":            {},
	"RECALLED":          {},
}

// Internal event types shielded from CUSTOMER/Public users
var internalEventTypes = map[string]struct{}{
	"PACKED":        {},
	"WAREHOUSE_IN":  {},
	"WAREHOUSE_OUT": {},
	"TRANSPORTED":   {},
}

func (s *traceService) SearchTimeline(ctx context.Context, req *request.TraceSearchRequest, userRole string, clientIP string, userAgent string, userID *uuid.UUID) (*response.TraceSearchResponse, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	if len(code) < 3 || len(code) > 100 {
		return nil, apperror.NewBadRequest("Vui lòng nhập mã sản phẩm hoặc số Serial hợp lệ để tìm kiếm")
	}

	// Parse date filters
	var fromDate, toDate *time.Time
	if req.FromDate != "" {
		t, err := time.Parse(time.RFC3339, req.FromDate)
		if err != nil {
			return nil, apperror.NewBadRequest("Invalid query parameters. Date must follow ISO-8601 format.")
		}
		fromDate = &t
	}
	if req.ToDate != "" {
		t, err := time.Parse(time.RFC3339, req.ToDate)
		if err != nil {
			return nil, apperror.NewBadRequest("Invalid query parameters. Date must follow ISO-8601 format.")
		}
		toDate = &t
	}

	if fromDate != nil && toDate != nil && fromDate.After(*toDate) {
		return nil, apperror.NewBadRequest("Invalid date range. Start date cannot be after end date.")
	}

	// Parse event type filter
	var filterEventTypes []string
	if req.EventTypes != "" {
		parts := strings.Split(req.EventTypes, ",")
		for _, part := range parts {
			part = strings.ToUpper(strings.TrimSpace(part))
			if part == "" {
				continue
			}
			if _, ok := validEventTypes[part]; !ok {
				return nil, apperror.NewBadRequest("Invalid event type filter. Supported values: PRODUCED, WAREHOUSE_IN, WAREHOUSE_OUT, SALE, REGISTERED, WARRANTY_CLAIM, WARRANTY_RESOLVED, RECALLED.")
			}
			filterEventTypes = append(filterEventTypes, part)
		}
	}

	// Fetch product item
	item, err := s.repo.FindProductItemByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, &apperror.AppError{
			Code:    apperror.CodeNotFound,
			Message: "Product item not found. Please verify your product code and try again.",
			Status:  http.StatusNotFound,
		}
	}

	// Apply security restrictions for CUSTOMER/Public users
	isPrivileged := userRole == "ADMIN" || userRole == "STAFF" || userRole == "DEALER"
	if !isPrivileged {
		// Public & Customers only see distributing/public milestones
		if len(filterEventTypes) > 0 {
			// Clean internal event types requested
			var sanitized []string
			for _, et := range filterEventTypes {
				if _, isInternal := internalEventTypes[et]; !isInternal {
					sanitized = append(sanitized, et)
				}
			}
			filterEventTypes = sanitized
		} else {
			// Query with all event types except internal ones
			for et := range validEventTypes {
				if _, isInternal := internalEventTypes[et]; !isInternal {
					filterEventTypes = append(filterEventTypes, et)
				}
			}
		}
	}

	// Query events
	events, err := s.repo.FindEvents(ctx, item.ItemID, item.BatchID, fromDate, toDate, filterEventTypes)
	if err != nil {
		return nil, err
	}

	// Build timeline DTOs
	var timelineDTOs []response.TimelineEventDTO
	for _, ev := range events {
		// Map DB event type to "RECALLED" for consistency if it's RECALL
		evType := ev.EventType
		if evType == "RECALL" {
			evType = "RECALLED"
		}
		timelineDTOs = append(timelineDTOs, response.TimelineEventDTO{
			EventID:     ev.EventID.String(),
			EventType:   evType,
			Title:       ev.Title,
			Description: ev.Description,
			Location:    ev.Location,
			Timestamp:   ev.Timestamp.Format(time.RFC3339),
		})
	}

	// Audit Logging
	action := "PUBLIC_SEARCH_TIMELINE"
	if req.FromDate != "" || req.ToDate != "" {
		action = "FILTER_TIMELINE_BY_DATE"
	} else if req.EventTypes != "" {
		action = "FILTER_TIMELINE_BY_TYPE"
	}

	auditPayload := map[string]any{
		"searchedCode":       code,
		"ipAddress":          clientIP,
		"userAgent":          userAgent,
		"timestamp":          time.Now().Format(time.RFC3339),
		"fromDate":           req.FromDate,
		"toDate":             req.ToDate,
		"eventTypesFiltered": filterEventTypes,
	}

	_ = s.auditService.Log(ctx, userID, action, "ProductItem", item.ItemID, nil, auditPayload)

	// Format final response DTO
	resp := &response.TraceSearchResponse{
		Timeline: timelineDTOs,
	}

	// Determine warning
	if item.Status == "RECALLED" {
		resp.Warning = "WARNING: This product belongs to a recalled batch. Please do not consume."
	}

	hasFilters := req.FromDate != "" || req.ToDate != "" || req.EventTypes != ""
	if hasFilters {
		var activeTypes []string
		if req.EventTypes != "" {
			activeTypes = filterEventTypes
		}
		var fromPtr *string
		if req.FromDate != "" {
			fromPtr = &req.FromDate
		}
		var toPtr *string
		if req.ToDate != "" {
			toPtr = &req.ToDate
		}
		resp.FilterApplied = &response.FilterAppliedDTO{
			FromDate:   fromPtr,
			ToDate:     toPtr,
			EventTypes: activeTypes,
		}
		count := len(timelineDTOs)
		resp.MatchedEventsCount = &count
	} else {
		resp.ProductItem = &response.ProductItemDTO{
			ItemID:       item.ItemID.String(),
			ItemCode:     item.ItemCode,
			SerialNumber: item.SerialNumber,
			Status:       item.Status,
			ProductName:  item.ProductName,
			ThumbnailURL: item.ThumbnailURL,
		}
	}

	return resp, nil
}

func (s *traceService) ExportPDF(ctx context.Context, req *request.PDFExportRequest, userID *uuid.UUID) (*response.ExportJobResponse, []byte, error) {
	// Query product item
	item, err := s.repo.FindProductItemByCode(ctx, req.ProductItemID.String())
	if err != nil {
		return nil, nil, err
	}
	if item == nil {
		return nil, nil, &apperror.AppError{
			Code:    apperror.CodeNotFound,
			Message: "Product item not found",
			Status:  http.StatusNotFound,
		}
	}

	// Fetch timeline events (privileged user gets all logs)
	events, err := s.repo.FindEvents(ctx, item.ItemID, item.BatchID, nil, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	if len(events) == 0 {
		return nil, nil, apperror.NewBadRequest("Cannot export PDF for a product with an empty timeline.")
	}

	var auditLogs []repositories.AuditLogDetail
	if req.IncludeAuditLogs {
		logs, err := s.repo.FindAuditLogs(ctx, item.ItemID, item.BatchID)
		if err != nil {
			return nil, nil, err
		}
		auditLogs = logs
	}

	// Synchronous generation for small timeline (< 10 events)
	if len(events) < 10 {
		pdfBytes, err := s.generatePDFBytes(item, events, auditLogs, req.Theme)
		if err != nil {
			return nil, nil, err
		}
		return nil, pdfBytes, nil
	}

	// Asynchronous generation
	jobID := uuid.New().String()
	jobKey := "trace_export_job:" + jobID

	jobStatus := map[string]string{
		"jobId":                jobID,
		"status":               "PROCESSING",
		"estimatedTimeSeconds": "5",
	}
	statusJSON, _ := json.Marshal(jobStatus)

	if s.redisClient != nil {
		_ = s.redisClient.Set(ctx, jobKey, string(statusJSON), 24*time.Hour).Err()
	}

	// Process asynchronously
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		pdfBytes, err := s.generatePDFBytes(item, events, auditLogs, req.Theme)
		if err != nil {
			s.updateJobError(bgCtx, jobKey, jobID, "PDF generation failed")
			return
		}

		// Ensure directory exists
		_ = os.MkdirAll("storage/temp", 0755)
		fileName := fmt.Sprintf("ProductJourney-%s-%d.pdf", item.SerialNumber, time.Now().Unix())
		filePath := "storage/temp/" + fileName

		if err := os.WriteFile(filePath, pdfBytes, 0644); err != nil {
			s.updateJobError(bgCtx, jobKey, jobID, "Storage write failed")
			return
		}

		downloadURL := fmt.Sprintf("%sstorage/temp/%s", s.baseURL, fileName)

		// Update job state
		completedStatus := map[string]string{
			"jobId":       jobID,
			"status":      "COMPLETED",
			"downloadUrl": downloadURL,
		}
		completedJSON, _ := json.Marshal(completedStatus)
		if s.redisClient != nil {
			_ = s.redisClient.Set(bgCtx, jobKey, string(completedJSON), 24*time.Hour)
		}

		// Publish event to RabbitMQ
		if s.pub != nil {
			event := types.Event{
				EventID:       uuid.NewString(),
				EventType:     "trace.exported",
				EventVersion:  "1.0",
				Timestamp:     time.Now().UTC(),
				Producer:      "go-core-service",
				CorrelationID: uuid.NewString(),
				Payload: map[string]interface{}{
					"userId":      userID,
					"format":      "PDF",
					"downloadUrl": downloadURL,
					"fileName":    fileName,
				},
			}
			_ = s.pub.Publish(event)
		}

		// Audit Log
		_ = s.auditService.Log(bgCtx, userID, "EXPORT_TRACE_TIMELINE_PDF", "ProductItem", item.ItemID, nil, map[string]any{
			"downloadUrl": downloadURL,
		})
	}()

	return &response.ExportJobResponse{
		JobID:                jobID,
		Status:               "PROCESSING",
		EstimatedTimeSeconds: 5,
	}, nil, nil
}

func (s *traceService) ExportExcel(ctx context.Context, req *request.ExcelExportRequest, userID *uuid.UUID) (*response.ExportJobResponse, []byte, error) {
	// Either ProductItemID or BatchID is required
	if req.ProductItemID == nil && req.BatchID == nil {
		return nil, nil, apperror.NewBadRequest("productItemId or batchId is required")
	}

	var fromDate, toDate *time.Time
	if req.FromDate != "" {
		t, err := time.Parse(time.RFC3339, req.FromDate)
		if err != nil {
			return nil, nil, apperror.NewBadRequest("Invalid date format fromDate")
		}
		fromDate = &t
	}
	if req.ToDate != "" {
		t, err := time.Parse(time.RFC3339, req.ToDate)
		if err != nil {
			return nil, nil, apperror.NewBadRequest("Invalid date format toDate")
		}
		toDate = &t
	}

	var productItems []*repositories.ProductItemDetail
	if req.ProductItemID != nil {
		item, err := s.repo.FindProductItemByCode(ctx, req.ProductItemID.String())
		if err != nil {
			return nil, nil, err
		}
		if item == nil {
			return nil, nil, &apperror.AppError{
				Code:    apperror.CodeNotFound,
				Message: "Product item not found",
				Status:  http.StatusNotFound,
			}
		}
		productItems = append(productItems, item)
	} else if req.BatchID != nil {
		// Query all product items in this batch from DB
		items, err := s.repo.FindProductItemsByBatchID(ctx, *req.BatchID)
		if err != nil {
			return nil, nil, err
		}
		productItems = items
	}

	if len(productItems) == 0 {
		return nil, nil, apperror.NewBadRequest("No product items found for export.")
	}

	// Check total events limit
	var totalEvents int
	var dataset []ItemWithEvents

	for _, item := range productItems {
		events, err := s.repo.FindEvents(ctx, item.ItemID, item.BatchID, fromDate, toDate, nil)
		if err != nil {
			return nil, nil, err
		}
		totalEvents += len(events)
		dataset = append(dataset, ItemWithEvents{Item: item, Events: events})
	}

	if totalEvents > 50000 {
		return nil, nil, apperror.NewBadRequest("Dataset too large. Please narrow your date range filter.")
	}

	// Synchronous download for single item with small dataset
	if req.ProductItemID != nil && totalEvents < 10 {
		xlsxBytes, err := s.generateExcelBytes(dataset)
		if err != nil {
			return nil, nil, err
		}
		return nil, xlsxBytes, nil
	}

	// Asynchronous generation
	jobID := uuid.New().String()
	jobKey := "trace_export_job:" + jobID

	jobStatus := map[string]string{
		"jobId":                jobID,
		"status":               "PROCESSING",
		"estimatedTimeSeconds": "3",
	}
	statusJSON, _ := json.Marshal(jobStatus)
	if s.redisClient != nil {
		_ = s.redisClient.Set(ctx, jobKey, string(statusJSON), 24*time.Hour)
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		xlsxBytes, err := s.generateExcelBytes(dataset)
		if err != nil {
			s.updateJobError(bgCtx, jobKey, jobID, "Excel generation failed")
			return
		}

		_ = os.MkdirAll("storage/temp", 0755)
		serial := "BatchExport"
		if req.ProductItemID != nil && len(productItems) > 0 {
			serial = productItems[0].SerialNumber
		}
		fileName := fmt.Sprintf("ProductJourney-%s-%d.xlsx", serial, time.Now().Unix())
		filePath := "storage/temp/" + fileName

		if err := os.WriteFile(filePath, xlsxBytes, 0644); err != nil {
			s.updateJobError(bgCtx, jobKey, jobID, "Storage write failed")
			return
		}

		downloadURL := fmt.Sprintf("%sstorage/temp/%s", s.baseURL, fileName)

		completedStatus := map[string]string{
			"jobId":       jobID,
			"status":      "COMPLETED",
			"downloadUrl": downloadURL,
		}
		completedJSON, _ := json.Marshal(completedStatus)
		if s.redisClient != nil {
			_ = s.redisClient.Set(bgCtx, jobKey, string(completedJSON), 24*time.Hour)
		}

		if s.pub != nil {
			event := types.Event{
				EventID:       uuid.NewString(),
				EventType:     "trace.exported",
				EventVersion:  "1.0",
				Timestamp:     time.Now().UTC(),
				Producer:      "go-core-service",
				CorrelationID: uuid.NewString(),
				Payload: map[string]interface{}{
					"userId":      userID,
					"format":      "EXCEL",
					"downloadUrl": downloadURL,
					"fileName":    fileName,
				},
			}
			_ = s.pub.Publish(event)
		}

		// Log audit
		var targetID uuid.UUID
		if req.ProductItemID != nil {
			targetID = *req.ProductItemID
		} else if req.BatchID != nil {
			targetID = *req.BatchID
		}
		_ = s.auditService.Log(bgCtx, userID, "EXPORT_TRACE_TIMELINE_EXCEL", "ProductItem", targetID, nil, map[string]any{
			"downloadUrl": downloadURL,
		})
	}()

	return &response.ExportJobResponse{
		JobID:                jobID,
		Status:               "PROCESSING",
		EstimatedTimeSeconds: 3,
	}, nil, nil
}

func (s *traceService) updateJobError(ctx context.Context, jobKey string, jobID string, message string) {
	errStatus := map[string]string{
		"jobId":  jobID,
		"status": "FAILED",
		"error":  message,
	}
	errJSON, _ := json.Marshal(errStatus)
	if s.redisClient != nil {
		_ = s.redisClient.Set(ctx, jobKey, string(errJSON), 24*time.Hour)
	}
}

func (s *traceService) generatePDFBytes(item *repositories.ProductItemDetail, events []repositories.TimelineEvent, auditLogs []repositories.AuditLogDetail, theme string) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Theme color settings
	primaryR, primaryG, primaryB := 20, 40, 80 // Classic Navy default
	if theme == "WARM_MINIMAL" {
		primaryR, primaryG, primaryB = 140, 80, 50 // Warm Orange/Brown
	}

	// 1. Header (Logo & Brand)
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(primaryR, primaryG, primaryB)
	pdf.CellFormat(0, 10, "PRODUCTTRACE AI", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "I", 10)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 5, "Official Product Origin & Life-cycle Certificate", "", 1, "L", false, 0, "")

	// Horizontal Rule
	pdf.SetDrawColor(primaryR, primaryG, primaryB)
	pdf.SetLineWidth(0.5)
	pdf.Line(15, 32, 195, 32)
	pdf.Ln(10)

	// 2. Product Information Block
	pdf.SetFillColor(245, 245, 245)
	pdf.Rect(15, 37, 180, 40, "F")

	pdf.SetTextColor(50, 50, 50)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetXY(18, 40)
	pdf.Cell(0, 6, "PRODUCT INFORMATION")
	pdf.Ln(7)

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetX(18)
	pdf.CellFormat(40, 6, "Product Name:", "", 0, "", false, 0, "")
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(120, 6, item.ProductName, "", 1, "", false, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetX(18)
	pdf.CellFormat(40, 6, "Item Code / QR:", "", 0, "", false, 0, "")
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(120, 6, item.ItemCode, "", 1, "", false, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	pdf.SetX(18)
	pdf.CellFormat(40, 6, "Serial Number:", "", 0, "", false, 0, "")
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(120, 6, item.SerialNumber, "", 1, "", false, 0, "")

	pdf.Ln(12)

	// 3. Timeline Title
	pdf.SetTextColor(primaryR, primaryG, primaryB)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.Cell(0, 8, "LIFE-CYCLE TIMELINE DISCOVERY")
	pdf.Ln(10)

	// 4. Draw Timeline
	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(80, 80, 80)

	timelineX := 25.0
	startY := pdf.GetY()

	// Draw vertical line background
	lineEndY := startY + float64(len(events)*24) - 10.0
	pdf.SetDrawColor(200, 200, 200)
	pdf.SetLineWidth(1.0)
	pdf.Line(timelineX, startY, timelineX, lineEndY)

	for i, ev := range events {
		currentY := startY + float64(i*24)
		pdf.SetXY(timelineX-2.5, currentY)

		// Draw circle indicator
		pdf.SetDrawColor(primaryR, primaryG, primaryB)
		pdf.SetFillColor(primaryR, primaryG, primaryB)
		pdf.Circle(timelineX, currentY+2, 2.5, "FD")

		// Write event details
		pdf.SetXY(timelineX+6, currentY)
		pdf.SetFont("Helvetica", "B", 10)
		pdf.SetTextColor(primaryR, primaryG, primaryB)
		pdf.Cell(0, 4, ev.Title)

		pdf.Ln(4)
		pdf.SetX(timelineX + 6)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(100, 100, 100)
		dateStr := ev.Timestamp.Format("2006-01-02 15:04:05 (UTC)")
		locStr := ev.Location
		if locStr == "" {
			locStr = "N/A"
		}
		pdf.Cell(0, 4, fmt.Sprintf("%s | Location: %s", dateStr, locStr))

		pdf.Ln(4)
		pdf.SetX(timelineX + 6)
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(60, 60, 60)
		pdf.MultiCell(150, 4, ev.Description, "", "", false)
	}

	pdf.SetY(lineEndY + 15)

	// 5. System Checksum
	h := sha256.New()
	h.Write([]byte(item.ItemCode + item.SerialNumber))
	for _, ev := range events {
		h.Write([]byte(ev.EventID.String() + ev.Title))
	}
	hashSum := hex.EncodeToString(h.Sum(nil))

	pdf.SetFont("Helvetica", "I", 7)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(0, 4, fmt.Sprintf("System digital watermarking hash verification checksum: %s", hashSum), "", 1, "C", false, 0, "")

	// 6. Appendix: Audit Logs
	if len(auditLogs) > 0 {
		pdf.AddPage()
		pdf.SetTextColor(primaryR, primaryG, primaryB)
		pdf.SetFont("Helvetica", "B", 12)
		pdf.Cell(0, 10, "APPENDIX: SYSTEM AUDIT HISTORY LOGS")
		pdf.Ln(10)

		// Draw table headers
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(230, 230, 230)
		pdf.SetTextColor(50, 50, 50)
		pdf.CellFormat(40, 7, "Timestamp (UTC)", "1", 0, "C", true, 0, "")
		pdf.CellFormat(30, 7, "User", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 7, "Action", "1", 0, "C", true, 0, "")
		pdf.CellFormat(25, 7, "Entity", "1", 0, "C", true, 0, "")
		pdf.CellFormat(60, 7, "Change Details", "1", 1, "C", true, 0, "")

		pdf.SetFont("Helvetica", "", 8)
		for _, log := range auditLogs {
			// Format summary
			details := fmt.Sprintf("Action: %s on %s", log.Action, log.Entity)
			if log.NewData != "" && len(log.NewData) < 150 {
				details = log.NewData
			}
			pdf.CellFormat(40, 6, log.CreatedAt.Format("2006-01-02 15:04:05"), "1", 0, "L", false, 0, "")
			pdf.CellFormat(30, 6, log.UserEmail, "1", 0, "L", false, 0, "")
			pdf.CellFormat(25, 6, log.Action, "1", 0, "C", false, 0, "")
			pdf.CellFormat(25, 6, log.Entity, "1", 0, "C", false, 0, "")
			pdf.CellFormat(60, 6, details, "1", 1, "L", false, 0, "")
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *traceService) generateExcelBytes(dataset []ItemWithEvents) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	// 1. General sheet creation
	sheet1 := "General Info"
	_ = f.SetSheetName("Sheet1", sheet1)

	// Set Title styling
	_ = f.SetCellValue(sheet1, "A1", "PRODUCTTRACE AI - DISCOVERY REPORT")
	_ = f.SetCellValue(sheet1, "A2", "Generated on: "+time.Now().Format("2006-01-02 15:04:05 UTC"))

	// Style headers
	styleHeader, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF", Size: 11},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"1A4D2E"}, Pattern: 1}, // Dark Green
	})

	styleZebra, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"F4F9F4"}, Pattern: 1}, // Pale Green/Gray
	})

	_ = f.SetCellValue(sheet1, "A4", "Mã sản phẩm (Item Code)")
	_ = f.SetCellValue(sheet1, "B4", "Số Serial (Serial)")
	_ = f.SetCellValue(sheet1, "C4", "Tên sản phẩm (Name)")
	_ = f.SetCellValue(sheet1, "D4", "Trạng thái (Status)")
	_ = f.SetRowHeight(sheet1, 4, 25)
	_ = f.SetColWidth(sheet1, "A", "D", 25)

	_ = f.SetCellStyle(sheet1, "A4", "D4", styleHeader)

	for i, data := range dataset {
		row := 5 + i
		_ = f.SetCellValue(sheet1, fmt.Sprintf("A%d", row), data.Item.ItemCode)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("B%d", row), data.Item.SerialNumber)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("C%d", row), data.Item.ProductName)
		_ = f.SetCellValue(sheet1, fmt.Sprintf("D%d", row), data.Item.Status)
		if i%2 == 1 {
			_ = f.SetCellStyle(sheet1, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), styleZebra)
		}
	}

	// 2. Detailed logs sheet
	sheet2 := "Timeline Log"
	_, _ = f.NewSheet(sheet2)

	headers := []string{"STT", "Mã sản phẩm", "Số Serial", "Event ID", "Loại sự kiện", "Tiêu đề", "Mô tả", "Địa điểm", "Người thực hiện", "Thời gian (UTC)"}
	for i, h := range headers {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetCellValue(sheet2, colName+"1", h)
	}
	_ = f.SetRowHeight(sheet2, 1, 25)
	_ = f.SetCellStyle(sheet2, "A1", "J1", styleHeader)

	currentRow := 2
	stt := 1
	for _, item := range dataset {
		for _, ev := range item.Events {
			_ = f.SetCellValue(sheet2, fmt.Sprintf("A%d", currentRow), stt)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("B%d", currentRow), item.Item.ItemCode)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("C%d", currentRow), item.Item.SerialNumber)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("D%d", currentRow), ev.EventID.String())
			_ = f.SetCellValue(sheet2, fmt.Sprintf("E%d", currentRow), ev.EventType)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("F%d", currentRow), ev.Title)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("G%d", currentRow), ev.Description)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("H%d", currentRow), ev.Location)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("I%d", currentRow), ev.ActorName)
			_ = f.SetCellValue(sheet2, fmt.Sprintf("J%d", currentRow), ev.Timestamp.Format("2006-01-02 15:04:05"))

			if stt%2 == 0 {
				_ = f.SetCellStyle(sheet2, fmt.Sprintf("A%d", currentRow), fmt.Sprintf("J%d", currentRow), styleZebra)
			}
			currentRow++
			stt++
		}
	}

	// Auto-fit column widths
	s.autoFitColumns(f, sheet1)
	s.autoFitColumns(f, sheet2)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *traceService) autoFitColumns(f *excelize.File, sheet string) {
	cols, _ := f.GetCols(sheet)
	for i, col := range cols {
		maxLen := 0
		for _, val := range col {
			if len(val) > maxLen {
				maxLen = len(val)
			}
		}
		colName, _ := excelize.ColumnNumberToName(i + 1)
		width := float64(maxLen + 3)
		if width < 10 {
			width = 10
		}
		_ = f.SetColWidth(sheet, colName, colName, width)
	}
}
