package response

import (
	"time"

	"github.com/google/uuid"
)

// BatchStatusResponse là DTO trả về sau khi cập nhật status của Batch.
type BatchStatusResponse struct {
	ID        uuid.UUID `json:"id"`
	BatchCode string    `json:"batch_code"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}
