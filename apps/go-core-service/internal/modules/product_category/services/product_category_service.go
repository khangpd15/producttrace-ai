package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
)

type ProductCategoryService interface {
	CreateCategory(ctx context.Context, req request.CreateCategoryRequest) (*entities.ProductCategory, error)
}

type productCategoryService struct {
	categoryRepo repositories.ProductCategoryRepository
}

func NewProductCategoryService(categoryRepo repositories.ProductCategoryRepository) ProductCategoryService {
	return &productCategoryService{categoryRepo: categoryRepo}
}

func (s *productCategoryService) CreateCategory(ctx context.Context, req request.CreateCategoryRequest) (*entities.ProductCategory, error) {
	// Validate name unique
	exists, err := s.categoryRepo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check category name"))
	}
	if exists {
		return nil, apperror.NewConflict("Category name already exists")
	}

	// Validate code unique
	exists, err = s.categoryRepo.ExistsByCode(ctx, req.Code)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check category code"))
	}
	if exists {
		return nil, apperror.NewConflict("Category code already exists")
	}

	// Validate parent tồn tại nếu có
	if req.ParentID != nil {
		exists, err = s.categoryRepo.ExistsByID(ctx, *req.ParentID)
		if err != nil {
			return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check parent category"))
		}
		if !exists {
			return nil, apperror.NewNotFound("parent category")
		}
	}

	category := &entities.ProductCategory{
		ID:          uuid.New(),
		Name:        req.Name,
		Code:        &req.Code,
		ParentID:    req.ParentID,
		Description: req.Description,
		IsActive:    req.Status == "ACTIVE",
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to create category"))
	}

	return category, nil
}
