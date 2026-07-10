package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/entity"
)

type WarrantyClaimResponse struct {
	ID                       uuid.UUID                  `json:"id"`
	ClaimNumber              string                     `json:"claim_number"`
	ProductItemID            uuid.UUID                  `json:"product_id"`
	IssueTitle               string                     `json:"issue_title"`
	IssueDescription         string                     `json:"issue_description"`
	ContactPhone             string                     `json:"contact_phone"`
	ContactEmail             *string                    `json:"contact_email,omitempty"`
	PreferredServiceCenterID *uuid.UUID                 `json:"preferred_service_center_id,omitempty"`
	Attachments              []string                   `json:"attachments,omitempty"`
	Status                   entity.WarrantyClaimStatus `json:"status"`
	CreatedAt                time.Time                  `json:"created_at"`
	UpdatedAt                time.Time                  `json:"updated_at"`
}
