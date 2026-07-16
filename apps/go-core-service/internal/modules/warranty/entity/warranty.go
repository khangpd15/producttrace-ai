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
	ID                uuid.UUID      `json:"id"`
	ProductItemID     uuid.UUID      `json:"product_item_id"`
	OwnerID           *uuid.UUID     `json:"owner_id"`
	ItemCode          string         `json:"itemCode"`
	ItemName          string         `json:"itemName"`
	SerialNumber      string         `json:"serialNumber"`
	OwnerName         string         `json:"ownerName"`
	OwnerEmail        string         `json:"ownerEmail"`
	WarrantyCode      string         `json:"warrantyCode"`
	PolicyName        string         `json:"policyName"`
	PolicyDescription string         `json:"policyDescription"`
	DurationMonths    int            `json:"durationMonths"`
	Status            WarrantyStatus `json:"status"`
	StartDate         time.Time      `json:"startDate"`
	EndDate           time.Time      `json:"endDate"`
	InvoiceNumber     string         `json:"invoiceNumber"`
	Note              string         `json:"note"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
}

func (Warranty) TableName() string {
	return "warranties"
}
