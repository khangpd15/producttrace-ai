package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/entities"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/repositories"
	categoryRepos "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_category/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"gorm.io/gorm"
)

type AttributeService interface {
	CreateAttribute(ctx context.Context, req request.CreateAttributeRequest) (*entities.Attribute, error)
	UpdateAttribute(ctx context.Context, id uuid.UUID, req request.UpdateAttributeRequest) (*entities.Attribute, error)
	GetAttributeByID(ctx context.Context, id uuid.UUID) (*entities.Attribute, error)
	ListAttributes(ctx context.Context, req request.ListAttributeRequest) ([]entities.Attribute, int64, error)
	DeleteAttribute(ctx context.Context, id uuid.UUID) error
}

type attributeService struct {
	attrRepo     repositories.AttributeRepository
	categoryRepo categoryRepos.ProductCategoryRepository
}

func NewAttributeService(attrRepo repositories.AttributeRepository, categoryRepo categoryRepos.ProductCategoryRepository) AttributeService {
	return &attributeService{
		attrRepo:     attrRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *attributeService) CreateAttribute(ctx context.Context, req request.CreateAttributeRequest) (*entities.Attribute, error) {
	// Check category exists
	categoryExists, err := s.categoryRepo.ExistsByID(ctx, req.CategoryID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check category"))
	}
	if !categoryExists {
		return nil, apperror.NewNotFound("category")
	}

	// Check code unique in category
	codeExists, err := s.attrRepo.ExistsByCodeAndCategory(ctx, req.Code, req.CategoryID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check attribute code"))
	}
	if codeExists {
		return nil, apperror.NewConflict("Attribute code already exists in this category")
	}

	attr := &entities.Attribute{
		ID:         uuid.New(),
		CategoryID: req.CategoryID,
		Code:       req.Code,
		Label:      req.Label,
		CreatedAt:  time.Now(),
		IsDeleted:  false,
	}

	if err := s.attrRepo.Create(ctx, attr); err != nil {
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to create attribute"))
	}

	return attr, nil
}

func (s *attributeService) UpdateAttribute(ctx context.Context, id uuid.UUID, req request.UpdateAttributeRequest) (*entities.Attribute, error) {
	attr, err := s.attrRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound("attribute")
		}
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to find attribute"))
	}

	if req.Code != nil && *req.Code != attr.Code {
		// Check code unique in category excluding current ID
		codeExists, err := s.attrRepo.ExistsByCodeAndCategoryExcludeID(ctx, *req.Code, attr.CategoryID, id)
		if err != nil {
			return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check attribute code"))
		}
		if codeExists {
			return nil, apperror.NewConflict("Attribute code already exists in this category")
		}
		attr.Code = *req.Code
	}

	if req.Label != nil {
		attr.Label = *req.Label
	}

	if err := s.attrRepo.Update(ctx, attr); err != nil {
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to update attribute"))
	}

	return attr, nil
}

func (s *attributeService) GetAttributeByID(ctx context.Context, id uuid.UUID) (*entities.Attribute, error) {
	attr, err := s.attrRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound("attribute")
		}
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to find attribute"))
	}
	return attr, nil
}

func (s *attributeService) ListAttributes(ctx context.Context, req request.ListAttributeRequest) ([]entities.Attribute, int64, error) {
	offset := (req.Page - 1) * req.Limit

	var categoryID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, 0, apperror.NewBadRequest("Invalid category ID format")
		}
		categoryID = &id
	}

	return s.attrRepo.FindAll(ctx, categoryID, req.Search, offset, req.Limit)
}

func (s *attributeService) DeleteAttribute(ctx context.Context, id uuid.UUID) error {
	_, err := s.attrRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NewNotFound("attribute")
		}
		return apperror.Wrap(err, apperror.NewInternal("Failed to find attribute"))
	}

	// Check if in use in attribute_values
	inUse, err := s.attrRepo.ExistsInAttributeValues(ctx, id)
	if err != nil {
		return apperror.Wrap(err, apperror.NewInternal("Failed to check if attribute is in use"))
	}
	if inUse {
		return apperror.NewConflict("Attribute is in use by variants and cannot be deleted")
	}

	if err := s.attrRepo.SoftDelete(ctx, id); err != nil {
		return apperror.Wrap(err, apperror.NewInternal("Failed to delete attribute"))
	}

	return nil
}
