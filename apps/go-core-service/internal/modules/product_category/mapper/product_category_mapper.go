package mapper

import (
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/dto/response"
    "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/entities"
)


func ToCategoryResponse(category *entities.ProductCategory) *response.CategoryResponse {
    children := make([]response.CategoryResponse, len(category.Children))
    for i, child := range category.Children {
        children[i] = *ToCategoryResponse(&child)
    }

    return &response.CategoryResponse{
        ID:          category.ID,
        Name:        category.Name,
        Code:        category.Code,
        ParentID:    category.ParentID,
        Description: category.Description,
        IsActive:    category.IsActive,
        CreatedAt:   category.CreatedAt,
        UpdatedAt:   category.UpdatedAt,
        Children:    children,
    }
}

func ToListCategoryResponse(categories []entities.ProductCategory, total int64, page, limit int) *response.ListCategoryResponse {
    data := make([]response.CategoryResponse, len(categories))
    for i, c := range categories {
        data[i] = *ToCategoryResponse(&c)
    }

    totalPages := int(total) / limit
    if int(total)%limit != 0 {
        totalPages++
    }

    return &response.ListCategoryResponse{
        Data:       data,
        Total:      total,
        Page:       page,
        Limit:      limit,
        TotalPages: totalPages,
    }
}