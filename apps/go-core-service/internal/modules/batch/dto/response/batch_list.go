package response

import (
	"time"

	"github.com/google/uuid"
)

type BatchListResponse struct {
	BatchCode  string     `json:"batch_code"`
	ProductID  uuid.UUID  `json:"product_id"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiryDate *time.Time `json:"expiry_date"`
}

func NewBatchListResponse(batchCode string, productID uuid.UUID, status string, createdAt time.Time, expiryDate *time.Time) *BatchListResponse {
	return &BatchListResponse{
		BatchCode:  batchCode,
		ProductID:  productID,
		Status:     status,
		CreatedAt:  createdAt,
		ExpiryDate: expiryDate,
	}
}
