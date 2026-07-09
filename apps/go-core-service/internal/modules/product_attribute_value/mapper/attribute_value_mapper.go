package mapper

import (
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_attribute_value/entities"
)

func ToAttributeValueResponse(val *entities.AttributeValue) response.AttributeValueResponse {
	return response.AttributeValueResponse{
		ID:               val.ID,
		ProductVariantID: val.ProductVariantID,
		AttributeID:      val.AttributeID,
		Label:            val.Label,
		ValueText:        val.ValueText,
		ValueNumber:      val.ValueNumber,
		ValueBoolean:     val.ValueBoolean,
		CreatedAt:        val.CreatedAt,
	}
}

func ToListAttributeValueResponse(vals []entities.AttributeValue, total int64, page, limit int) *response.ListAttributeValueResponse {
	data := make([]response.AttributeValueResponse, len(vals))
	for i, v := range vals {
		data[i] = ToAttributeValueResponse(&v)
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return &response.ListAttributeValueResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
