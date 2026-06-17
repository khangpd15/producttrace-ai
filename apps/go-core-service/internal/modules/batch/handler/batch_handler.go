package handler

import (
	"github.com/gin-gonic/gin"
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
