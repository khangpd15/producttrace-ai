package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/audit/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/audit/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/audit_log"
	pkg_response "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type AuditHandler struct {
	auditService audit_log.AuditLogService
}

func NewAuditHandler(auditService audit_log.AuditLogService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// GetLogs handles both Audit Log API and Audit Search API
func (h *AuditHandler) GetLogs(c *gin.Context) {
	var req request.GetAuditLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 10
	}

	logs, total, err := h.auditService.GetLogs(c.Request.Context(), req.Action, req.Entity, req.FromDate, req.ToDate, page, limit)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	var resData []response.AuditLogResponse
	for _, log := range logs {
		resData = append(resData, response.AuditLogResponse{
			ID:        log.ID,
			UserID:    log.UserID,
			Action:    log.Action,
			Entity:    log.Entity,
			EntityID:  log.EntityID,
			OldData:   log.OldData,
			NewData:   log.NewData,
			CreatedAt: log.CreatedAt,
		})
	}
	
	if resData == nil {
		resData = []response.AuditLogResponse{}
	}

	c.JSON(200, pkg_response.ResponseSuccess("Audit logs retrieved successfully", map[string]interface{}{
		"data":  resData,
		"total": total,
		"page":  page,
		"limit": limit,
	}))
}
