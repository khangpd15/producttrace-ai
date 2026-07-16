package entity

import (
	"time"

	"github.com/google/uuid"
)

type ClaimStatus string

const (
	ClaimStatusPending   ClaimStatus = "PENDING"
	ClaimStatusApproved  ClaimStatus = "APPROVED"
	ClaimStatusRejected  ClaimStatus = "REJECTED"
	ClaimStatusInRepair  ClaimStatus = "IN_REPAIR"
	ClaimStatusCompleted ClaimStatus = "COMPLETED"
)

type WarrantyClaim struct {
	ID               uuid.UUID   `json:"id" gorm:"column:id;type:uuid;primaryKey"`
	WarrantyID       uuid.UUID   `json:"warranty_id" gorm:"column:warranty_id;type:uuid;not null;index"`
	ProductItemID    uuid.UUID   `json:"product_item_id" gorm:"column:product_item_id;type:uuid;not null;index"`
	CustomerName     string      `json:"customerName" gorm:"column:customer_name;type:varchar;not null"`
	CustomerPhone    string      `json:"customerPhone" gorm:"column:customer_phone;type:varchar;not null"`
	CustomerEmail    string      `json:"customerEmail" gorm:"column:customer_email;type:varchar"`
	IssueTitle       string      `json:"issueTitle" gorm:"column:issue_title;type:varchar;not null"`
	IssueDescription string      `json:"issueDescription" gorm:"column:issue_description;type:text"`
	Status           ClaimStatus `json:"status" gorm:"column:status;type:varchar;default:PENDING;index"`
	ResolutionNote   string      `json:"resolutionNote" gorm:"column:resolution_note;type:text"`
	CreatedAt        time.Time   `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time   `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (WarrantyClaim) TableName() string {
	return "warranty_claims"
}
