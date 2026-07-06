package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/services"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/utils"
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

	currentUserIDStr := utils.GetCurrentUserID(c)

	if currentUserIDStr == "" {
		apperror.HandleError(c, apperror.NewInternal("fail to get current user id"))
		return
	}

	currentUserID, err := uuid.Parse(currentUserIDStr)

	if err == nil {
		apperror.HandleError(c, apperror.NewInternal("fail to parse current user id"))
		return
	}

	result, err := hb.service.CreateBatch(c.Request.Context(), &req, currentUserID)
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

// UpdateBatchStatus xử lý PATCH /batches/:batch_id/status.
// Handler chỉ parse params, bind body, gọi service và trả response.
func (hb *BatchHandler) UpdateBatchStatus(c *gin.Context) {
	batchIDStr := c.Param("batch_id")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewBadRequest("invalid batch_id: must be a valid UUID"))
		return
	}

	var req request.UpdateBatchStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	// Lấy current user ID từ context (có thể nil với system actions).
	var userID *uuid.UUID
	if idStr := utils.GetCurrentUserID(c); idStr != "" {
		if parsed, parseErr := uuid.Parse(idStr); parseErr == nil {
			userID = &parsed
		}
	}

	result, svcErr := hb.service.UpdateBatchStatus(c.Request.Context(), batchID, &req, userID)
	if svcErr != nil {
		apperror.HandleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("batch status updated successfully", result))
}

// DeleteBatch xử lý DELETE /batches/:batch_id.
// Trả 200 OK khi soft-delete thành công.
func (hb *BatchHandler) DeleteBatch(c *gin.Context) {
	batchIDStr := c.Param("batch_id")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewBadRequest("invalid batch_id: must be a valid UUID"))
		return
	}

	// Lấy current user ID từ context.
	var userID *uuid.UUID
	if idStr := utils.GetCurrentUserID(c); idStr != "" {
		if parsed, parseErr := uuid.Parse(idStr); parseErr == nil {
			userID = &parsed
		}
	}

	if svcErr := hb.service.DeleteBatch(c.Request.Context(), batchID, userID); svcErr != nil {
		apperror.HandleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("batch deleted successfully", nil))
}
