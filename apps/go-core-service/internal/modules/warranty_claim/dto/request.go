package dto

import "github.com/google/uuid"

type CreateWarrantyClaimRequest struct {
	ProductID                uuid.UUID `json:"product_id" binding:"required"`
	IssueTitle               string    `json:"issue_title" binding:"required,max=255"`
	IssueDescription         string    `json:"issue_description" binding:"required,max=5000"`
	ContactPhone             string    `json:"contact_phone" binding:"required,max=50"`
	ContactEmail             string    `json:"contact_email,omitempty" binding:"omitempty,email,max=255"`
	PreferredServiceCenterID string    `json:"preferred_service_center,omitempty" binding:"omitempty,uuid"`
	Attachments              []string  `json:"attachments,omitempty"`
}
