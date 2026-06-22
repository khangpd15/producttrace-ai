package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/services"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type BatchHandler struct {
	service services.BatchService
}

func NewBatchHandler(service services.BatchService) *BatchHandler {
	return &BatchHandler{
		service: service,
	}
}

func (hb *BatchHandler) GetBatchList(c *gin.Context) {
	batches, err := hb.service.GetBatchList(c.Request.Context())

	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("get batch list successfully", batches))
}

func (hb *BatchHandler) GetBatchDetail(c *gin.Context) {
	batchCode := c.Param("batch_code")
	if batchCode == "" {
		apperror.HandleError(c, apperror.NewBadRequest("batch_code is required"))
		return
	}

	detail, err := hb.service.GetBatchDetail(c.Request.Context(), batchCode)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("get batch detail successfully", detail))
}

func (hb *BatchHandler) CreateBatch(c *gin.Context) {
	var req request.CreateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	result, err := hb.service.CreateBatch(c.Request.Context(), &req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(201, response.ResponseSuccess("batch created successfully", result))
}

func (h *BatchHandler) ExportQR(c *gin.Context) {
	fmt.Println("FULL PATH:", c.FullPath())
	fmt.Println("batch_id:", c.Param("batch_id"))

	batchIDStr := c.Param("batch_id")

	if batchIDStr == "" {
		apperror.HandleError(c, apperror.NewBadRequest("missing batchID"))
		return
	}

	batchID, err := uuid.Parse(batchIDStr)

	if err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	pdfBytes, err := h.service.ExportBatchQR(
		c.Request.Context(),
		batchID,
	)

	if err != nil {
		return
	}

	c.Header(
		"Content-Disposition",
		`attachment; filename="batch_qr.pdf"`,
	)

	c.Data(
		http.StatusOK,
		"application/pdf",
		pdfBytes,
	)
}
