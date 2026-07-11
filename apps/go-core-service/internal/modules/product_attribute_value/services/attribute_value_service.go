package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/entities"
	valRepos "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/repositories"
	variantRepos "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/repositories"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"gorm.io/gorm"
)

type AttributeValueService interface {
	AssignAttributes(ctx context.Context, variantID uuid.UUID, req request.BulkCreateAttributeValuesRequest) ([]entities.AttributeValue, error)
	UpdateAttributeValue(ctx context.Context, id uuid.UUID, req request.UpdateAttributeValueRequest) (*entities.AttributeValue, error)
	DeleteAttributeValue(ctx context.Context, id uuid.UUID) error
	GetAttributeValueByID(ctx context.Context, id uuid.UUID) (*entities.AttributeValue, error)
	GetAttributeValuesByVariantID(ctx context.Context, variantID uuid.UUID) ([]entities.AttributeValue, error)
	ListAllAttributeValues(ctx context.Context, req request.ListAttributeValueRequest) ([]entities.AttributeValue, int64, error)
}

type attributeValueService struct {
	db          *gorm.DB
	valRepo     valRepos.AttributeValueRepository
	variantRepo variantRepos.ProductVariantRepository
	attrRepo    repositories.AttributeRepository
}

func NewAttributeValueService(
	db *gorm.DB,
	valRepo valRepos.AttributeValueRepository,
	variantRepo variantRepos.ProductVariantRepository,
	attrRepo repositories.AttributeRepository,
) AttributeValueService {
	return &attributeValueService{
		db:          db,
		valRepo:     valRepo,
		variantRepo: variantRepo,
		attrRepo:    attrRepo,
	}
}

func (s *attributeValueService) AssignAttributes(ctx context.Context, variantID uuid.UUID, req request.BulkCreateAttributeValuesRequest) ([]entities.AttributeValue, error) {
	// 1. Verify variant exists
	variantExists, err := s.variantRepo.ExistsByID(ctx, variantID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check product variant"))
	}
	if !variantExists {
		return nil, apperror.NewNotFound("product variant")
	}

	createdVals := make([]entities.AttributeValue, 0, len(req.Items))

	// Run in transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		txCtx := valRepos.InjectTx(ctx, tx)

		// To prevent duplicate attributes within the same batch list
		seenAttributes := make(map[uuid.UUID]bool)

		for _, item := range req.Items {
			// Check if duplicate attribute id in the request payload
			if seenAttributes[item.AttributeID] {
				return apperror.NewBadRequest("Duplicate attribute assignment in request")
			}
			seenAttributes[item.AttributeID] = true

			// 2. Verify attribute exists
			attrExists, err := s.attrRepo.ExistsByID(txCtx, item.AttributeID)
			if err != nil {
				return apperror.Wrap(err, apperror.NewInternal("Failed to check attribute"))
			}
			if !attrExists {
				return apperror.NewNotFound("attribute")
			}

			// 3. Check if variant already has value assigned for this attribute
			exists, err := s.valRepo.ExistsByVariantAndAttribute(txCtx, variantID, item.AttributeID)
			if err != nil {
				return apperror.Wrap(err, apperror.NewInternal("Failed to check attribute value association"))
			}
			if exists {
				return apperror.NewConflict("Attribute is already assigned to this variant")
			}

			val := entities.AttributeValue{
				ID:               uuid.New(),
				ProductVariantID: variantID,
				AttributeID:      item.AttributeID,
				Label:            item.Label,
				ValueText:        item.ValueText,
				ValueNumber:      item.ValueNumber,
				ValueBoolean:     item.ValueBoolean,
				CreatedAt:        time.Now(),
			}

			if err := s.valRepo.Create(txCtx, &val); err != nil {
				return apperror.Wrap(err, apperror.NewInternal("Failed to save attribute value"))
			}

			createdVals = append(createdVals, val)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdVals, nil
}

func (s *attributeValueService) UpdateAttributeValue(ctx context.Context, id uuid.UUID, req request.UpdateAttributeValueRequest) (*entities.AttributeValue, error) {
	val, err := s.valRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound("attribute value")
		}
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to find attribute value"))
	}

	if req.Label != nil {
		val.Label = *req.Label
	}
	if req.ValueText != nil {
		val.ValueText = req.ValueText
	}
	if req.ValueNumber != nil {
		val.ValueNumber = req.ValueNumber
	}
	if req.ValueBoolean != nil {
		val.ValueBoolean = req.ValueBoolean
	}

	if err := s.valRepo.Update(ctx, val); err != nil {
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to update attribute value"))
	}

	return val, nil
}

func (s *attributeValueService) DeleteAttributeValue(ctx context.Context, id uuid.UUID) error {
	exists, err := s.valRepo.ExistsByID(ctx, id)
	if err != nil {
		return apperror.Wrap(err, apperror.NewInternal("Failed to check attribute value"))
	}
	if !exists {
		return apperror.NewNotFound("attribute value")
	}

	if err := s.valRepo.Delete(ctx, id); err != nil {
		return apperror.Wrap(err, apperror.NewInternal("Failed to delete attribute value"))
	}

	return nil
}

func (s *attributeValueService) GetAttributeValueByID(ctx context.Context, id uuid.UUID) (*entities.AttributeValue, error) {
	val, err := s.valRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewNotFound("attribute value")
		}
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to find attribute value"))
	}
	return val, nil
}

func (s *attributeValueService) GetAttributeValuesByVariantID(ctx context.Context, variantID uuid.UUID) ([]entities.AttributeValue, error) {
	// Check variant exists first
	variantExists, err := s.variantRepo.ExistsByID(ctx, variantID)
	if err != nil {
		return nil, apperror.Wrap(err, apperror.NewInternal("Failed to check product variant"))
	}
	if !variantExists {
		return nil, apperror.NewNotFound("product variant")
	}

	return s.valRepo.FindByVariantID(ctx, variantID)
}

func (s *attributeValueService) ListAllAttributeValues(ctx context.Context, req request.ListAttributeValueRequest) ([]entities.AttributeValue, int64, error) {
	offset := (req.Page - 1) * req.Limit

	var variantID *uuid.UUID
	if req.VariantID != nil && *req.VariantID != "" {
		id, err := uuid.Parse(*req.VariantID)
		if err != nil {
			return nil, 0, apperror.NewBadRequest("Invalid variant ID format")
		}
		variantID = &id
	}

	var attributeID *uuid.UUID
	if req.AttributeID != nil && *req.AttributeID != "" {
		id, err := uuid.Parse(*req.AttributeID)
		if err != nil {
			return nil, 0, apperror.NewBadRequest("Invalid attribute ID format")
		}
		attributeID = &id
	}

	return s.valRepo.FindAll(ctx, variantID, attributeID, offset, req.Limit)
}
