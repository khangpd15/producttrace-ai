package entity

import (
	"time"

	"github.com/google/uuid"
)

type WarrantyStatus string

const (
	WarrantyStatusInactive  WarrantyStatus = "INACTIVE"
	WarrantyStatusPending   WarrantyStatus = "PENDING"
	WarrantyStatusActive    WarrantyStatus = "ACTIVE"
	WarrantyStatusExpired   WarrantyStatus = "EXPIRED"
	WarrantyStatusClaimed   WarrantyStatus = "CLAIMED"
	WarrantyStatusResolved  WarrantyStatus = "RESOLVED"
	WarrantyStatusRejected  WarrantyStatus = "REJECTED"
	WarrantyStatusCancelled WarrantyStatus = "CANCELLED"
)

type Warranty struct {
	ID                uuid.UUID      `json:"id" gorm:"column:id;type:uuid;primaryKey"`
	ProductItemID     uuid.UUID      `json:"product_item_id" gorm:"column:product_item_id;type:uuid;not null;index"`
	OwnerID           *uuid.UUID     `json:"owner_id" gorm:"column:owner_id;type:uuid;index"`
	ItemCode          string         `json:"itemCode" gorm:"column:item_code;type:varchar;not null;index"`
	ItemName          string         `json:"itemName" gorm:"column:item_name;type:varchar"`
	SerialNumber      string         `json:"serialNumber" gorm:"column:serial_number;type:varchar;not null;index"`
	OwnerName         string         `json:"ownerName" gorm:"column:owner_name;type:varchar;not null"`
	OwnerEmail        string         `json:"ownerEmail" gorm:"column:owner_email;type:varchar"`
	WarrantyCode      string         `json:"warrantyCode" gorm:"column:warranty_code;type:varchar;uniqueIndex;not null"`
	PolicyName        string         `json:"policyName" gorm:"column:policy_name;type:varchar;not null"`
	PolicyDescription string         `json:"policyDescription" gorm:"column:policy_description;type:text"`
	DurationMonths    int            `json:"durationMonths" gorm:"column:duration_months;type:int"`
	Status            WarrantyStatus `json:"status" gorm:"column:status;type:varchar;default:ACTIVE;index"`
	StartDate         time.Time      `json:"startDate" gorm:"column:start_date"`
	EndDate           time.Time      `json:"endDate" gorm:"column:end_date"`
	InvoiceNumber     string         `json:"invoiceNumber" gorm:"column:invoice_number;type:varchar"`
	Note              string         `json:"note" gorm:"column:note;type:text"`
	CreatedAt         time.Time      `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time      `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (Warranty) TableName() string {
	return "warranties"
}
