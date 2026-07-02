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
		// Status phải được set tường minh vì GORM gửi empty string "" cho field không được gán,
		// vi phạm CHECK constraint chk_product_items_status trên DB.
		// "IN_STOCK" là trạng thái khởi đầu của mọi item vừa được tạo trong batch.
		Status: "IN_STOCK",
	}
}
