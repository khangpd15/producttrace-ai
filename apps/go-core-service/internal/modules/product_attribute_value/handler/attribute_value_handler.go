package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/dto/request"
	dto "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/mapper"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/services"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type AttributeValueHandler struct {
	valService services.AttributeValueService
}

func NewAttributeValueHandler(valService services.AttributeValueService) *AttributeValueHandler {
	return &AttributeValueHandler{valService: valService}
}

func handleError(c *gin.Context, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus(), response.ResponseError(appErr.Message, nil))
		return
	}
	c.JSON(http.StatusInternalServerError, response.ResponseError("Internal server error", nil))
}

func (h *AttributeValueHandler) AssignAttributes(c *gin.Context) {
	variantIDStr := c.Param("variant_id")
	variantID, err := uuid.Parse(variantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError("Invalid variant id", nil))
		return
	}

	var req request.BulkCreateAttributeValuesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
		return
	}

	vals, err := h.valService.AssignAttributes(c.Request.Context(), variantID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	res := make([]dto.AttributeValueResponse, len(vals))
	for i, v := range vals {
		res[i] = mapper.ToAttributeValueResponse(&v)
	}

	c.JSON(http.StatusCreated, response.ResponseSuccess("Attributes assigned successfully", res))
}

func (h *AttributeValueHandler) GetAttributeValuesByVariantID(c *gin.Context) {
	variantIDStr := c.Param("id")
	variantID, err := uuid.Parse(variantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError("Invalid variant id", nil))
		return
	}

	vals, err := h.valService.GetAttributeValuesByVariantID(c.Request.Context(), variantID)
	if err != nil {
		handleError(c, err)
		return
	}

	res := make([]dto.AttributeValueResponse, len(vals))
	for i, v := range vals {
		res[i] = mapper.ToAttributeValueResponse(&v)
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("OK", res))
}

func (h *AttributeValueHandler) UpdateAttributeValue(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError("Invalid attribute value id", nil))
		return
	}

	var req request.UpdateAttributeValueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
		return
	}

	val, err := h.valService.UpdateAttributeValue(c.Request.Context(), id, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Attribute value updated successfully", mapper.ToAttributeValueResponse(val)))
}

func (h *AttributeValueHandler) DeleteAttributeValue(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError("Invalid attribute value id", nil))
		return
	}

	err = h.valService.DeleteAttributeValue(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Attribute value deleted successfully", nil))
}

func (h *AttributeValueHandler) ListAllAttributeValues(c *gin.Context) {
	var req request.ListAttributeValueRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
		return
	}

	vals, total, err := h.valService.ListAllAttributeValues(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("OK", mapper.ToListAttributeValueResponse(vals, total, req.Page, req.Limit)))
}

func (h *AttributeValueHandler) GetAttributeValueByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError("Invalid attribute value id", nil))
		return
	}

	val, err := h.valService.GetAttributeValueByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("OK", mapper.ToAttributeValueResponse(val)))
}
