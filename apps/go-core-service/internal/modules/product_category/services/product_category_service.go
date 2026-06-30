package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"gorm.io/gorm"
)

type ProductCategoryService interface {
	CreateCategory(ctx context.Context, req request.CreateCategoryRequest) (*entities.ProductCategory, error)
	UpdateCategory(ctx context.Context, id uuid.UUID, req request.UpdateCategoryRequest) (*entities.ProductCategory, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) error
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*entities.ProductCategory, error)
	GetAllCategories(ctx context.Context, filter request.ListCategoryRequest) ([]entities.ProductCategory, int64, error)
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

func (s *productCategoryService) UpdateCategory(ctx context.Context, id uuid.UUID, req request.UpdateCategoryRequest) (*entities.ProductCategory, error) {
	// Kiểm tra category tồn tại
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound("category")
		}
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to find category"))
	}

	// Validate name unique (trừ chính nó)
	if req.Name != nil {
		exists, err := s.categoryRepo.ExistsByNameExcludeID(ctx, *req.Name, id)
		if err != nil {
			return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check category name"))
		}
		if exists {
			return nil, apperror.NewConflict("Category name already exists")
		}
		category.Name = *req.Name
	}

	// Validate code unique (trừ chính nó)
	if req.Code != nil {
		exists, err := s.categoryRepo.ExistsByCodeExcludeID(ctx, *req.Code, id)
		if err != nil {
			return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check category code"))
		}
		if exists {
			return nil, apperror.NewConflict("Category code already exists")
		}
		category.Code = req.Code
	}

	// Validate parent tồn tại nếu có
	if req.ParentID != nil {
		exists, err := s.categoryRepo.ExistsByID(ctx, *req.ParentID)
		if err != nil {
			return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check parent category"))
		}
		if !exists {
			return nil, apperror.NewNotFound("parent category")
		}
		category.ParentID = req.ParentID
	}

	if req.Description != nil {
		category.Description = req.Description
	}

	if req.Status != nil {
		category.IsActive = *req.Status == "ACTIVE"
	}

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to update category"))
	}

	return category, nil
}

func (s *productCategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	// Check tồn tại
	_, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return apperror.NewNotFound("category")
	}

	// Check đang được dùng
	inUse, err := s.categoryRepo.IsInUse(ctx, id)
	if err != nil {
		return apperror.Wrap(err, apperror.NewInternal("Failed to check category usage"))
	}
	if inUse {
		return apperror.NewConflict("Category is in use")
	}

	// Soft delete
	if err := s.categoryRepo.SoftDelete(ctx, id); err != nil {
		return apperror.Wrap(err, apperror.NewInternal("Failed to delete category"))
	}

	return nil
}

func (s *productCategoryService) GetCategoryByID(ctx context.Context, id uuid.UUID) (*entities.ProductCategory, error) {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.NewNotFound("category")
	}
	return category, nil
}

func (s *productCategoryService) GetAllCategories(ctx context.Context, filter request.ListCategoryRequest) ([]entities.ProductCategory, int64, error) {
	repoFilter := repositories.CategoryFilter{
		Search: filter.Search,
		Status: filter.Status,
		Page:   filter.Page,
		Limit:  filter.Limit,
	}
	return s.categoryRepo.FindAll(ctx, repoFilter)
}
