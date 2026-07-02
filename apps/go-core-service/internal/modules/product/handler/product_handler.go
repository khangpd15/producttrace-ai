package handler

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/dto/request"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/mapper"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/services"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type ProductHandler struct {
    productService services.ProductService
}

func NewProductHandler(productService services.ProductService) *ProductHandler {
    return &ProductHandler{productService: productService}
}

func handleError(c *gin.Context, err error) {
    var appErr *apperror.AppError
    if errors.As(err, &appErr) {
        c.JSON(appErr.HTTPStatus(), response.ResponseError(appErr.Message, nil))
        return
    }
    c.JSON(http.StatusInternalServerError, response.ResponseError("Internal server error", nil))
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
    var req request.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
        return
    }

    // Mock user ID tạm thời
    createdBy := uuid.New()

    product, err := h.productService.CreateProduct(c.Request.Context(), req, createdBy)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusCreated, response.ResponseSuccess("Product created successfully", mapper.ToProductResponse(product)))
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, response.ResponseError("Invalid product id", nil))
        return
    }

    var req request.UpdateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
        return
    }

    product, err := h.productService.UpdateProduct(c.Request.Context(), id, req)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, response.ResponseSuccess("Product updated successfully", mapper.ToProductResponse(product)))
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, response.ResponseError("Invalid product id", nil))
        return
    }

    product, err := h.productService.GetProductByID(c.Request.Context(), id)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, response.ResponseSuccess("OK", mapper.ToProductResponse(product)))
}

func (h *ProductHandler) GetAllProducts(c *gin.Context) {
    var filter request.ListProductRequest
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

    products, total, err := h.productService.GetAllProducts(c.Request.Context(), filter)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, response.ResponseSuccess("OK", mapper.ToListProductResponse(products, total, filter.Page, filter.Limit)))
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, response.ResponseError("Invalid product id", nil))
        return
    }

    if err := h.productService.DeleteProduct(c.Request.Context(), id); err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, response.ResponseSuccess("Product deleted successfully", nil))
}