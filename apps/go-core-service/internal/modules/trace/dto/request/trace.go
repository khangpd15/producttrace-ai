package request

import (
	"github.com/google/uuid"
)

// TraceSearchRequest represents the query parameters for GET /api/v1/trace/search.
type TraceSearchRequest struct {
	Code       string `form:"code" binding:"required,min=3,max=100"`
	FromDate   string `form:"fromDate"`
	ToDate     string `form:"toDate"`
	EventTypes string `form:"eventTypes"`
}

// PDFExportRequest represents the JSON request body for POST /api/v1/trace/export/pdf.
type PDFExportRequest struct {
	ProductItemID    uuid.UUID `json:"productItemId" binding:"required"`
	Theme            string    `json:"theme"` // WARM_MINIMAL, CLASSIC_NAVY
	IncludeAuditLogs bool      `json:"includeAuditLogs"`
}

// ExcelExportRequest represents the JSON request body for POST /api/v1/trace/export/excel.
type ExcelExportRequest struct {
	ProductItemID    *uuid.UUID `json:"productItemId"`
	BatchID          *uuid.UUID `json:"batchId"`
	FromDate         string     `json:"fromDate"`
	ToDate           string     `json:"toDate"`
}
