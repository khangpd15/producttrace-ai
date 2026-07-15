package entity

import (
	"time"

	"github.com/google/uuid"
)

type OwnershipStatus string

const (
	OwnershipStatusPending     OwnershipStatus = "PENDING"
	OwnershipStatusActive      OwnershipStatus = "ACTIVE"
	OwnershipStatusTransferred OwnershipStatus = "TRANSFERRED"
	OwnershipStatusRevoked     OwnershipStatus = "REVOKED"
)

type Ownership struct {
	ID            uuid.UUID       `json:"id" gorm:"column:id;type:uuid;primaryKey"`
	ProductItemID uuid.UUID       `json:"product_item_id" gorm:"column:product_item_id;type:uuid;not null"`
	OwnerID       uuid.UUID       `json:"owner_id" gorm:"column:owner_id;type:uuid;not null"`
	Status        OwnershipStatus `json:"status" gorm:"column:status;default:ACTIVE"`
	OwnershipType string          `json:"ownership_type" gorm:"column:ownership_type;default:PRIMARY"`
	OwnedAt       time.Time       `json:"owned_at" gorm:"column:owned_at"`
	EndedAt       *time.Time      `json:"ended_at" gorm:"column:ended_at"`
	CreatedAt     time.Time       `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time       `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func NewOwnership(productItemID uuid.UUID, ownerID uuid.UUID) *Ownership {
	return &Ownership{
		ID:            uuid.New(),
		ProductItemID: productItemID,
		OwnerID:       ownerID,
		Status:        OwnershipStatusActive,
		OwnershipType: "PRIMARY",
		OwnedAt:       time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
