package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type WarrantyClaimStatus string

const (
	WarrantyClaimStatusOpen       WarrantyClaimStatus = "OPEN"
	WarrantyClaimStatusProcessing WarrantyClaimStatus = "PROCESSING"
	WarrantyClaimStatusRejected   WarrantyClaimStatus = "REJECTED"
	WarrantyClaimStatusApproved   WarrantyClaimStatus = "APPROVED"
)

type WarrantyClaim struct {
	ID                       uuid.UUID           `json:"id" gorm:"column:id;type:uuid;primaryKey"`
	ClaimNumber              string              `json:"claim_number" gorm:"column:claim_number;type:varchar;uniqueIndex;not null"`
	ProductItemID            uuid.UUID           `json:"product_item_id" gorm:"column:product_item_id;type:uuid;not null"`
	IssueTitle               string              `json:"issue_title" gorm:"column:issue_title;type:varchar(255);not null"`
	IssueDescription         string              `json:"issue_description" gorm:"column:issue_description;type:text;not null"`
	ContactPhone             string              `json:"contact_phone" gorm:"column:contact_phone;type:varchar(50);not null"`
	ContactEmail             *string             `json:"contact_email" gorm:"column:contact_email;type:varchar(255)"`
	PreferredServiceCenterID *uuid.UUID          `json:"preferred_service_center_id" gorm:"column:preferred_service_center_id;type:uuid"`
	AttachmentsJSON          datatypes.JSON      `json:"attachments" gorm:"column:attachments_json;type:jsonb"`
	Status                   WarrantyClaimStatus `json:"status" gorm:"column:status;type:varchar(50);default:OPEN;not null"`
	CreatedBy                uuid.UUID           `json:"created_by" gorm:"column:created_by;type:uuid;not null"`
	CreatedAt                time.Time           `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt                time.Time           `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (WarrantyClaim) TableName() string {
	return "warranty_claims"
}
