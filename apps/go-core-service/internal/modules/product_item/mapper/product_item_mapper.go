package mapper

import (
	"github.com/google/uuid"
	batchDTO "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/batch/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/product_item/entities"
)

func CreateProductItemRequestToEntity(newID uuid.UUID, itemCode string, verificationToken string, serialNumber string, batch *batchDTO.BatchCreateResponse) *entities.ProductItem {
	return &entities.ProductItem{
		ID:                newID,
		BatchID:           batch.ID,
		VariantID:         batch.VariantID,
		ItemCode:          itemCode,
		VerificationToken: verificationToken,
		SerialNumber:      serialNumber,
	}
}
