package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/mapper"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/services"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type AttributeHandler struct {
	attrService services.AttributeService
}

func NewAttributeHandler(attrService services.AttributeService) *AttributeHandler {
	return &AttributeHandler{attrService: attrService}
}

func handleError(c *gin.Context, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus(), response.ResponseError(appErr.Message, nil))
		return
	}
	c.JSON(http.StatusInternalServerError, response.ResponseError("Internal server error", nil))
}

func (h *AttributeHandler) CreateAttribute(c *gin.Context) {
	var req request.CreateAttributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
		return
	}

	attr, err := h.attrService.CreateAttribute(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.ResponseSuccess("Attribute created successfully", mapper.ToAttributeResponse(attr)))
}

func (h *AttributeHandler) UpdateAttribute(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError("Invalid attribute id", nil))
		return
	}

	var req request.UpdateAttributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
		return
	}

	attr, err := h.attrService.UpdateAttribute(c.Request.Context(), id, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Attribute updated successfully", mapper.ToAttributeResponse(attr)))
}

func (h *AttributeHandler) GetAttributeByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError("Invalid attribute id", nil))
		return
	}

	attr, err := h.attrService.GetAttributeByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("OK", mapper.ToAttributeResponse(attr)))
}

func (h *AttributeHandler) ListAttributes(c *gin.Context) {
	var req request.ListAttributeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
		return
	}

	attrs, total, err := h.attrService.ListAttributes(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	attrList := mapper.ToListAttributeResponse(attrs, total, req.Page, req.Limit)
	c.JSON(http.StatusOK, response.ResponseSuccess("OK", attrList.Data))
}

func (h *AttributeHandler) DeleteAttribute(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError("Invalid attribute id", nil))
		return
	}

	err = h.attrService.DeleteAttribute(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Attribute deleted successfully", nil))
}
