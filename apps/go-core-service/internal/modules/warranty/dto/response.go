package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/entity"
)

type WarrantyResponse struct {
	ID                uuid.UUID             `json:"id"`
	ProductItemID     uuid.UUID             `json:"product_item_id"`
	OwnerID           *uuid.UUID            `json:"owner_id,omitempty"`
	WarrantyCode      string                `json:"warranty_code"`
	PolicyName        string                `json:"policy_name"`
	PolicyDescription string                `json:"policy_description"`
	DurationMonths    int                   `json:"duration_months"`
	Status            entity.WarrantyStatus `json:"status"`
	StartDate         *time.Time            `json:"start_date,omitempty"`
	EndDate           *time.Time            `json:"end_date,omitempty"`
	ActivatedAt       *time.Time            `json:"activated_at,omitempty"`
	InvoiceNumber     string                `json:"invoice_number"`
	InvoiceURL        string                `json:"invoice_url,omitempty"`
	Note              string                `json:"note,omitempty"`
	ItemCode          string                `json:"item_code,omitempty"`
	SerialNumber      string                `json:"serial_number,omitempty"`
	OwnerName         string                `json:"owner_name,omitempty"`
	OwnerEmail        string                `json:"owner_email,omitempty"`
	CreatedAt         time.Time             `json:"created_at"`
	UpdatedAt         time.Time             `json:"updated_at"`
}

type WarrantyListResponse struct {
	Items []*WarrantyResponse `json:"items"`
	Total int64               `json:"total"`
}
