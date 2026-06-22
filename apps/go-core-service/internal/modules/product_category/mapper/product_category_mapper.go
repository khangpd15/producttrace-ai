package mapper

import (
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/dto/response"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/entities"
)

func ToCategoryResponse(category *entities.ProductCategory) *response.CategoryResponse {
    return &response.CategoryResponse{
        ID:          category.ID,
        Name:        category.Name,
        Code:        category.Code,
        ParentID:    category.ParentID,
        Description: category.Description,
        IsActive:    category.IsActive,
        CreatedAt:   category.CreatedAt,
        UpdatedAt:   category.UpdatedAt,
    }
}