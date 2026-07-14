package response

import (
	"time"

	"github.com/google/uuid"
)

// BatchCreateResponse là DTO trả về sau khi tạo lô hàng thành công.
// Chỉ trả về các field cần thiết để confirm với client.
type BatchCreateResponse struct {
	ID        uuid.UUID `json:"id"`
	BatchCode string    `json:"batch_code"`
	VariantID uuid.UUID `json:"variant_id"`
	Quantity  int       `json:"quantity"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
