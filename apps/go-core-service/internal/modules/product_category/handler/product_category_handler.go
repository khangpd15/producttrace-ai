package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/services"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/mapper"
)

type ProductCategoryHandler struct {
	categoryService services.ProductCategoryService
}

func NewProductCategoryHandler(categoryService services.ProductCategoryService) *ProductCategoryHandler {
	return &ProductCategoryHandler{categoryService: categoryService}
}

func handleError(c *gin.Context, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus(), response.ResponseError(appErr.Message, nil))
		return
	}
	c.JSON(http.StatusInternalServerError, response.ResponseError("Internal server error", nil))
}

func (h *ProductCategoryHandler) CreateCategory(c *gin.Context) {
    var req request.CreateCategoryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
        return
    }

    category, err := h.categoryService.CreateCategory(c.Request.Context(), req)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusCreated, response.ResponseSuccess("Category created successfully", mapper.ToCategoryResponse(category)))
}