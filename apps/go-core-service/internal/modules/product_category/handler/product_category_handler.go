package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/mapper"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/services"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
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

func (h *ProductCategoryHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError("Invalid category id", nil))
		return
	}

	var req request.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
		return
	}

	category, err := h.categoryService.UpdateCategory(c.Request.Context(), id, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Category updated successfully", mapper.ToCategoryResponse(category)))
}

func (h *ProductCategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError("Invalid category id", nil))
		return
	}

	if err := h.categoryService.DeleteCategory(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Category deleted successfully", nil))
}

func (h *ProductCategoryHandler) GetCategoryByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError("Invalid category id", nil))
		return
	}

	category, err := h.categoryService.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("OK", mapper.ToCategoryResponse(category)))
}

func (h *ProductCategoryHandler) GetAllCategories(c *gin.Context) {
	var filter request.ListCategoryRequest
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

	categories, total, err := h.categoryService.GetAllCategories(c.Request.Context(), filter)
	if err != nil {
		handleError(c, err)
		return
	}

	categoryList := mapper.ToListCategoryResponse(categories, total, filter.Page, filter.Limit)
	c.JSON(http.StatusOK, response.ResponseSuccess("OK", categoryList.Data))
}
