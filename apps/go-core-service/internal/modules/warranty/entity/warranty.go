package entity

import (
	"time"

	"github.com/google/uuid"
)

type WarrantyStatus string

const (
	WarrantyStatusInactive  WarrantyStatus = "INACTIVE"
	WarrantyStatusActive    WarrantyStatus = "ACTIVE"
	WarrantyStatusExpired   WarrantyStatus = "EXPIRED"
	WarrantyStatusClaimed   WarrantyStatus = "CLAIMED"
	WarrantyStatusResolved  WarrantyStatus = "RESOLVED"
	WarrantyStatusRejected  WarrantyStatus = "REJECTED"
	WarrantyStatusCancelled WarrantyStatus = "CANCELLED"
)

type Warranty struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ProductItemID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"product_item_id"`
	OwnerID           *uuid.UUID     `gorm:"type:uuid" json:"owner_id"`
	WarrantyCode      string         `gorm:"type:varchar;uniqueIndex" json:"warranty_code"`
	PolicyName        string         `gorm:"type:varchar" json:"policy_name"`
	PolicyDescription string         `gorm:"type:text" json:"policy_description"`
	DurationMonths    int            `gorm:"type:integer" json:"duration_months"`
	Status            WarrantyStatus `gorm:"type:varchar;default:'INACTIVE'" json:"status"`
	StartDate         *time.Time     `json:"start_date"`
	EndDate           *time.Time     `json:"end_date"`
	ActivatedAt       *time.Time     `json:"activated_at"`
	InvoiceNumber     string         `gorm:"type:varchar" json:"invoice_number"`
	InvoiceURL        string         `gorm:"type:text" json:"invoice_url"`
	Note              string         `gorm:"type:text" json:"note"`
	ItemCode          string         `gorm:"-" json:"item_code,omitempty"`
	SerialNumber      string         `gorm:"-" json:"serial_number,omitempty"`
	OwnerName         string         `gorm:"-" json:"owner_name,omitempty"`
	OwnerEmail        string         `gorm:"-" json:"owner_email,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

func (Warranty) TableName() string {
	return "warranties"
}
