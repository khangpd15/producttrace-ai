package dto

import (
	"time"

	"github.com/google/uuid"
)

type RequestWarrantyActivationRequest struct {
	ProductItemID uuid.UUID `json:"product_item_id" binding:"required"`
	InvoiceNumber string    `json:"invoice_number" binding:"required,max=100"`
	InvoiceURL    string    `json:"invoice_url,omitempty" binding:"omitempty,url"`
	Note          string    `json:"note,omitempty" binding:"omitempty,max=500"`
}

type ConfirmWarrantyRequest struct {
	Decision          string `json:"decision" binding:"required,oneof=APPROVE REJECT"`
	PolicyName        string `json:"policy_name,omitempty" binding:"omitempty,max=100"`
	PolicyDescription string `json:"policy_description,omitempty" binding:"omitempty,max=1000"`
	DurationMonths    int    `json:"duration_months,omitempty" binding:"omitempty,min=1"`
	Note              string `json:"note,omitempty" binding:"omitempty,max=500"`
}

type UpdateWarrantyRequest struct {
	PolicyName        string     `json:"policy_name,omitempty"`
	PolicyDescription string     `json:"policy_description,omitempty"`
	DurationMonths    int        `json:"duration_months,omitempty"`
	StartDate         *time.Time `json:"start_date,omitempty"`
	EndDate           *time.Time `json:"end_date,omitempty"`
	Status            string     `json:"status,omitempty"`
	Note              string     `json:"note,omitempty"`
}

type GetWarrantyListRequest struct {
	Search string `form:"search"`
	Status string `form:"status"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}
