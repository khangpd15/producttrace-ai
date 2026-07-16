package dto

import "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/entity"

type WarrantyResponse struct {
	ID                string `json:"id"`
	ProductItemID     string `json:"product_item_id"`
	OwnerID           string `json:"owner_id,omitempty"`
	ItemCode          string `json:"itemCode"`
	ItemName          string `json:"itemName"`
	SerialNumber      string `json:"serialNumber"`
	OwnerName         string `json:"ownerName"`
	OwnerEmail        string `json:"ownerEmail"`
	WarrantyCode      string `json:"warrantyCode"`
	PolicyName        string `json:"policyName"`
	PolicyDescription string `json:"policyDescription"`
	DurationMonths    int    `json:"durationMonths"`
	Status            string `json:"status"`
	StartDate         string `json:"startDate"`
	EndDate           string `json:"endDate"`
	InvoiceNumber     string `json:"invoiceNumber"`
	Note              string `json:"note"`
	CreatedAt         string `json:"createdAt"`
	UpdatedAt         string `json:"updatedAt"`
}

func FromWarrantyEntity(w *entity.Warranty) WarrantyResponse {
	ownerID := ""
	if w.OwnerID != nil {
		ownerID = w.OwnerID.String()
	}
	return WarrantyResponse{
		ID:                w.ID.String(),
		ProductItemID:     w.ProductItemID.String(),
		OwnerID:           ownerID,
		ItemCode:          w.ItemCode,
		ItemName:          w.ItemName,
		SerialNumber:      w.SerialNumber,
		OwnerName:         w.OwnerName,
		OwnerEmail:        w.OwnerEmail,
		WarrantyCode:      w.WarrantyCode,
		PolicyName:        w.PolicyName,
		PolicyDescription: w.PolicyDescription,
		DurationMonths:    w.DurationMonths,
		Status:            string(w.Status),
		StartDate:         w.StartDate.Format("2006-01-02"),
		EndDate:           w.EndDate.Format("2006-01-02"),
		InvoiceNumber:     w.InvoiceNumber,
		Note:              w.Note,
		CreatedAt:         w.CreatedAt.Format("2006-01-02 15:04"),
		UpdatedAt:         w.UpdatedAt.Format("2006-01-02 15:04"),
	}
}

func FromWarrantyEntities(list []entity.Warranty) []WarrantyResponse {
	res := make([]WarrantyResponse, 0, len(list))
	for _, w := range list {
		res = append(res, FromWarrantyEntity(&w))
	}
	return res
}
