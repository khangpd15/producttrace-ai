package response

import (
	"time"

	"github.com/google/uuid"
)

// BatchHistoryActorDTO chứa thông tin người thực hiện thay đổi.
// JOIN từ bảng users qua user_id của audit_log.
type BatchHistoryActorDTO struct {
	UserID   uuid.UUID `json:"userId"`
	FullName string    `json:"fullName"`
	Role     string    `json:"role"`
}

// FieldDiffDTO là cấu trúc diff-view cho một trường thay đổi (BR-HIS-001).
// old = giá trị trước khi thay đổi, new = giá trị sau khi thay đổi.
type FieldDiffDTO struct {
	Old any `json:"old"`
	New any `json:"new"`
}

// BatchHistoryItemDTO là một dòng lịch sử thay đổi.
// changedFields là map động — key là tên trường, value là {old, new}.
type BatchHistoryItemDTO struct {
	LogID         uuid.UUID                `json:"logId"`
	Action        string                   `json:"action"`
	ChangedFields map[string]FieldDiffDTO  `json:"changedFields"`
	PerformedBy   *BatchHistoryActorDTO    `json:"performedBy"`
	IPAddress     string                   `json:"ipAddress"`
	CreatedAt     time.Time                `json:"createdAt"`
}

// GetBatchHistoryResponse là response DTO theo UC-P2-BATCH-06.
// Không có pagination vì tài liệu không yêu cầu metadata phân trang trong response.
type GetBatchHistoryResponse struct {
	BatchID   uuid.UUID             `json:"batchId"`
	BatchCode string                `json:"batchCode"`
	History   []BatchHistoryItemDTO `json:"history"`
}
