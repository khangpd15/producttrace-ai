package handler

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/dto/request"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/mapper"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/services"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type ProductVariantHandler struct {
    variantService services.ProductVariantService
}

func NewProductVariantHandler(variantService services.ProductVariantService) *ProductVariantHandler {
    return &ProductVariantHandler{variantService: variantService}
}

func handleError(c *gin.Context, err error) {
    var appErr *apperror.AppError
    if errors.As(err, &appErr) {
        c.JSON(appErr.HTTPStatus(), response.ResponseError(appErr.Message, nil))
        return
    }
    c.JSON(http.StatusInternalServerError, response.ResponseError("Internal server error", nil))
}

func (h *ProductVariantHandler) GetVariantByID(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, response.ResponseError("Invalid variant id", nil))
        return
    }

    variant, err := h.variantService.GetVariantByID(c.Request.Context(), id)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, response.ResponseSuccess("OK", mapper.ToVariantResponse(variant)))
}

func (h *ProductVariantHandler) GetVariantsByProductID(c *gin.Context) {
    productIDStr := c.Param("product_id")
    productID, err := uuid.Parse(productIDStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, response.ResponseError("Invalid product id", nil))
        return
    }

    var filter request.ListVariantRequest
    if err := c.ShouldBindQuery(&filter); err != nil {
        c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
        return
    }

    if filter.Page == 0 {
        filter.Page = 1
    }
    if filter.Limit == 0 {
        filter.Limit = 10
    }

    variants, total, err := h.variantService.GetVariantsByProductID(c.Request.Context(), productID, filter)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, response.ResponseSuccess("OK", mapper.ToListVariantResponse(variants, total)))
}

func (h *ProductVariantHandler) UpdateVariant(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, response.ResponseError("Invalid variant id", nil))
        return
    }

    var req request.UpdateVariantRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
        return
    }

    variant, err := h.variantService.UpdateVariant(c.Request.Context(), id, req)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, response.ResponseSuccess("Variant updated successfully", mapper.ToVariantResponse(variant)))
}

func (h *ProductVariantHandler) DeleteVariant(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, response.ResponseError("Invalid variant id", nil))
        return
    }

    if err := h.variantService.DeleteVariant(c.Request.Context(), id); err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, response.ResponseSuccess("Variant deleted successfully", nil))
}