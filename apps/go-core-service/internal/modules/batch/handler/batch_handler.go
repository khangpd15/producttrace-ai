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
	var req request.GetBatchListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	// Lấy role từ JWT context để service thực thi BR-FIL-002 (ẩn DRAFT với non-Admin).
	userRole := utils.GetCurrentRole(c)

	batches, err := hb.service.GetBatchList(c.Request.Context(), &req, userRole)

	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("get batch list successfully", batches))
}

// SearchBatch xử lý GET /api/v1/batches/search theo UC-P2-BATCH-03.
// Handler chịu trách nhiệm: parse query params → validate → gọi service → trả response.
// Không chứa business logic.
func (hb *BatchHandler) SearchBatch(c *gin.Context) {
	var req request.SearchBatchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	result, err := hb.service.SearchBatches(c.Request.Context(), &req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("search batches successfully", result))
}

func (hb *BatchHandler) GetBatchDetail(c *gin.Context) {
	batchCode := c.Param("id")
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

func (hb *BatchHandler) GetBatchEvents(c *gin.Context) {
	batchIDStr := c.Param("id")
	if batchIDStr == "" {
		apperror.HandleError(c, apperror.NewBadRequest("batch_id is required"))
		return
	}
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewBadRequest("invalid batch_id"))
		return
	}

	events, err := hb.service.GetBatchEvents(c.Request.Context(), batchID)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("get batch events successfully", events))
}

func (hb *BatchHandler) ExportBatch(c *gin.Context) {
	batchIDStr := c.Param("id")
	if batchIDStr == "" {
		apperror.HandleError(c, apperror.NewBadRequest("batch_id is required"))
		return
	}
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewBadRequest("invalid batch_id"))
		return
	}

	var req request.ExportBatchRequest
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
	if err != nil {
		apperror.HandleError(c, apperror.NewInternal("fail to parse current user id"))
		return
	}

	err = hb.service.ExportBatch(c.Request.Context(), batchID, &req, currentUserID)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("export batch successfully", nil))
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
	fmt.Println("currentUserIDStr:", currentUserIDStr)
	currentUserID, err := uuid.Parse(currentUserIDStr)
	fmt.Println("currentUserID:", currentUserID)

	if err != nil {
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
	fmt.Println("batch_id:", c.Param("id"))

	batchIDStr := c.Param("id")

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
	batchIDStr := c.Param("id")
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
	batchIDStr := c.Param("id")
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

// GetBatchHistory xử lý GET /api/v1/batches/:id/history theo UC-P2-BATCH-06.
// Handler chịu trách nhiệm: parse batchId (UUID), bind query params → gọi service → trả response.
func (hb *BatchHandler) GetBatchHistory(c *gin.Context) {
	batchIDStr := c.Param("id")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewBadRequest("ID lô sản xuất không đúng cấu trúc UUID"))
		return
	}

	var req request.GetBatchHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	// Lấy userID để service truyền vào event payload (non-blocking).
	var userID *uuid.UUID
	if idStr := utils.GetCurrentUserID(c); idStr != "" {
		if parsed, parseErr := uuid.Parse(idStr); parseErr == nil {
			userID = &parsed
		}
	}

	result, svcErr := hb.service.GetBatchHistory(c.Request.Context(), batchID, &req, userID)
	if svcErr != nil {
		apperror.HandleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("get batch history successfully", result))
}

// GetBatchProducts xử lý GET /api/v1/batches/:id/products theo UC-P2-BATCH-05.
// Handler chịu trách nhiệm: parse batchId (UUID), bind query params → gọi service → trả response.
func (hb *BatchHandler) GetBatchProducts(c *gin.Context) {
	batchIDStr := c.Param("id")
	batchID, err := uuid.Parse(batchIDStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewBadRequest("invalid batch_id: must be a valid UUID"))
		return
	}

	var req request.GetBatchProductsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	result, svcErr := hb.service.GetBatchProducts(c.Request.Context(), batchID, &req)
	if svcErr != nil {
		apperror.HandleError(c, svcErr)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("get batch products successfully", result))
}

