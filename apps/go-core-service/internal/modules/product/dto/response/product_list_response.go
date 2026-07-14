package response

import (
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type ProductListItemDTO struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	CategoryName  string    `json:"category_name"`
	VariantsCount int       `json:"variants_count"`
	BatchesCount  int       `json:"batches_count"`
	Status        *string   `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	ThumbnailURL  *string   `json:"thumbnail_url"`
}

type ProductListResponse struct {
	Items []ProductListItemDTO    `json:"items"`
	Meta  response.PaginationMeta `json:"meta"`
}
