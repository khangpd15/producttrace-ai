package mapper

import (
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/dto/request"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/entities"
)

func CreateProductItemRequestToEntity(newID uuid.UUID, itemCode string, verificationToken string, serialNumber string, req *request.CreateProductItemRequest) *entities.ProductItem {
	return &entities.ProductItem{
		ID:                newID,
		BatchID:           req.BatchID,
		VariantID:         req.VariantID,
		ItemCode:          itemCode,
		VerificationToken: verificationToken,
		SerialNumber:      serialNumber,
	}
}
