package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/trace/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/trace/services"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/utils"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type TraceHandler struct {
	service services.TraceService
}

func NewTraceHandler(service services.TraceService) *TraceHandler {
	return &TraceHandler{service: service}
}

func (h *TraceHandler) Search(c *gin.Context) {
	var req request.TraceSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation("Vui lòng nhập mã sản phẩm hoặc số Serial hợp lệ để tìm kiếm"))
		return
	}

	userRole := utils.GetCurrentRole(c)
	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	var userID *uuid.UUID
	if idStr := utils.GetCurrentUserID(c); idStr != "" {
		if u, err := uuid.Parse(idStr); err == nil {
			userID = &u
		}
	}

	res, err := h.service.SearchTimeline(c.Request.Context(), &req, userRole, clientIP, userAgent, userID)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Search timeline successfully", res))
}

func (h *TraceHandler) ExportPDF(c *gin.Context) {
	var req request.PDFExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	var userID *uuid.UUID
	if idStr := utils.GetCurrentUserID(c); idStr != "" {
		if u, err := uuid.Parse(idStr); err == nil {
			userID = &u
		}
	}

	job, pdfBytes, err := h.service.ExportPDF(c.Request.Context(), &req, userID)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	if pdfBytes != nil {
		fileName := fmt.Sprintf("ProductJourney-%s-%d.pdf", req.ProductItemID.String(), time.Now().Unix())
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
		c.Data(http.StatusOK, "application/pdf", pdfBytes)
		return
	}

	c.JSON(http.StatusAccepted, response.ResponseSuccess("PDF export job initiated.", job))
}

func (h *TraceHandler) ExportExcel(c *gin.Context) {
	var req request.ExcelExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	var userID *uuid.UUID
	if idStr := utils.GetCurrentUserID(c); idStr != "" {
		if u, err := uuid.Parse(idStr); err == nil {
			userID = &u
		}
	}

	job, xlsxBytes, err := h.service.ExportExcel(c.Request.Context(), &req, userID)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	if xlsxBytes != nil {
		serial := "Export"
		if req.ProductItemID != nil {
			serial = req.ProductItemID.String()
		}
		fileName := fmt.Sprintf("ProductJourney-%s-%d.xlsx", serial, time.Now().Unix())
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", xlsxBytes)
		return
	}

	c.JSON(http.StatusAccepted, response.ResponseSuccess("Excel export job initiated.", job))
}
