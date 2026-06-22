package response

import (
    "time"
    "github.com/google/uuid"
)

type CategoryResponse struct {
    ID          uuid.UUID  `json:"id"`
    Name        string     `json:"name"`
    Code        *string    `json:"code"`
    ParentID    *uuid.UUID `json:"parent_id"`
    Description *string    `json:"description"`
    IsActive    bool       `json:"is_active"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}