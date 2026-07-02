package dto

import (
	"time"

	"github.com/google/uuid"
)

// OwnershipHistoryItem đại diện cho 1 lần đăng ký / chuyển giao quyền sở hữu
type OwnershipHistoryItem struct {
	OwnershipID      uuid.UUID  `json:"ownership_id"`
	OwnerName        string     `json:"owner_name"`
	OwnerEmail       string     `json:"owner_email"`
	OwnerPhone       string     `json:"owner_phone"`
	Status           string     `json:"status"`
	RegistrationDate time.Time  `json:"registration_date"` // tương ứng owned_at trong DB
	EndedAt          *time.Time `json:"ended_at,omitempty"`
}

// OwnershipDetailRes là toàn bộ response cho View Ownership Detail (UC-P1-OWNER-02)
type OwnershipDetailRes struct {
	// --- Thông tin quyền sở hữu hiện tại ---
	OwnershipID      uuid.UUID `json:"ownership_id"`
	ProductID        uuid.UUID `json:"product_id"` // = product_item_id trong DB
	Status           string    `json:"status"`
	RegistrationDate time.Time `json:"registration_date"` // owned_at

	// --- Thông tin chủ sở hữu hiện tại (fetch từ User Module) ---
	OwnerID    uuid.UUID `json:"owner_id"`
	OwnerName  string    `json:"owner_name"`
	OwnerEmail string    `json:"owner_email"`
	OwnerPhone string    `json:"owner_phone"`

	// --- Thông tin sản phẩm (fetch từ Product Module) ---
	ProductName string `json:"product_name"`
	ProductSKU  string `json:"product_sku"`

	// --- Lịch sử đăng ký quyền sở hữu (AC-002) ---
	OwnershipHistory []OwnershipHistoryItem `json:"ownership_history"`
}
