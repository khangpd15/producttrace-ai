package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/services"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type ProductItemHandler struct {
	service services.ProductItemService
}

func NewProductItemHandler(service services.ProductItemService) *ProductItemHandler {
	return &ProductItemHandler{
		service: service,
	}
}

func (h *ProductItemHandler) GetProductItemList(c *gin.Context) {
	var req request.GetProductItemListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	result, err := h.service.GetProductItemList(c.Request.Context(), &req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("OK", result))
}

func (h *ProductItemHandler) GetProductItemDetail(c *gin.Context) {
	itemCode := c.Param("item_code")
	if itemCode == "" {
		apperror.HandleError(c, apperror.NewBadRequest("item_code is required"))
		return
	}

	result, err := h.service.GetProductItemDetail(c.Request.Context(), itemCode)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("OK", result))
}
