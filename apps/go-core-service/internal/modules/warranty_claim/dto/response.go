package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/entity"
)

type WarrantyClaimResponse struct {
	ID               uuid.UUID `json:"id"`
	WarrantyID       uuid.UUID `json:"warrantyId"`
	ProductItemID    uuid.UUID `json:"productItemId"`
	CustomerName     string    `json:"customerName"`
	CustomerPhone    string    `json:"customerPhone"`
	CustomerEmail    string    `json:"customerEmail"`
	IssueTitle       string    `json:"issueTitle"`
	IssueDescription string    `json:"issueDescription"`
	Status           string    `json:"status"`
	ResolutionNote   string    `json:"resolutionNote"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func FromEntity(e *entity.WarrantyClaim) WarrantyClaimResponse {
	return WarrantyClaimResponse{
		ID:               e.ID,
		WarrantyID:       e.WarrantyID,
		ProductItemID:    e.ProductItemID,
		CustomerName:     e.CustomerName,
		CustomerPhone:    e.CustomerPhone,
		CustomerEmail:    e.CustomerEmail,
		IssueTitle:       e.IssueTitle,
		IssueDescription: e.IssueDescription,
		Status:           string(e.Status),
		ResolutionNote:   e.ResolutionNote,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}
}

func FromEntities(claims []entity.WarrantyClaim) []WarrantyClaimResponse {
	var resp []WarrantyClaimResponse
	for _, c := range claims {
		resp = append(resp, FromEntity(&c))
	}
	return resp
}
