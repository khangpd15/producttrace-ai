package mapper

import (
	"encoding/json"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product/entities"
	attrValMapper "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/mapper"
	attrValResponse "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/dto/response"
	variantEntities "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_variant/entities"
)

func ToProductResponse(product *entities.Product) *response.ProductResponse {
	// Convert tags JSON → []string
	var tags []string
	if product.Tags != nil {
		json.Unmarshal(product.Tags, &tags)
	}

	// Convert metadata JSON → map
	var metadata map[string]interface{}
	if product.MetadataJSON != nil {
		json.Unmarshal(product.MetadataJSON, &metadata)
	}

	variants := make([]response.VariantResponse, len(product.Variants))
	for i, v := range product.Variants {
		variants[i] = ToVariantResponse(&v)
	}

	return &response.ProductResponse{
		ID:           product.ID,
		CategoryID:   product.CategoryID,
		Name:         product.Name,
		Slug:         product.Slug,
		Description:  product.Description,
		ThumbnailURL: product.ThumbnailURL,
		Tags:         tags,
		Metadata:     metadata,
		Status:       product.Status,
		CreatedBy:    product.CreatedBy,
		CreatedAt:    product.CreatedAt,
		UpdatedAt:    product.UpdatedAt,
		Variants:     variants,
	}
}

func ToVariantResponse(variant *variantEntities.ProductVariant) response.VariantResponse {
	// Convert ImagesJSON → []string
	var images []string
	if variant.ImagesJSON != nil {
		json.Unmarshal(variant.ImagesJSON, &images)
	}

	attributes := make([]attrValResponse.AttributeValueResponse, len(variant.AttributeValues))
	for i, v := range variant.AttributeValues {
		attributes[i] = attrValMapper.ToAttributeValueResponse(&v)
	}

	return response.VariantResponse{
		ID:         variant.ID,
		SKU:        variant.SKU,
		Name:       variant.Name,
		Barcode:    variant.Barcode,
		Price:      variant.Price,
		Currency:   variant.Currency,
		Images:     images,
		Status:     variant.Status,
		Attributes: attributes,
		CreatedAt:  variant.CreatedAt,
		UpdatedAt:  variant.UpdatedAt,
	}
}