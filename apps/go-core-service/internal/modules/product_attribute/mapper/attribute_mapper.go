package mapper

import (
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute/entities"
)

func ToAttributeResponse(attr *entities.Attribute) response.AttributeResponse {
	return response.AttributeResponse{
		ID:         attr.ID,
		CategoryID: attr.CategoryID,
		Code:       attr.Code,
		Label:      attr.Label,
		CreatedAt:  attr.CreatedAt,
	}
}

func ToListAttributeResponse(attrs []entities.Attribute, total int64, page, limit int) *response.ListAttributeResponse {
	data := make([]response.AttributeResponse, len(attrs))
	for i, a := range attrs {
		data[i] = ToAttributeResponse(&a)
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &response.ListAttributeResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
