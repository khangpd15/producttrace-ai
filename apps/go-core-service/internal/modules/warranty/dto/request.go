package dto

type CreateWarrantyRequest struct {
	ItemCode          string `json:"itemCode" binding:"required"`
	ItemName          string `json:"itemName"`
	SerialNumber      string `json:"serialNumber" binding:"required"`
	OwnerName         string `json:"ownerName" binding:"required"`
	OwnerEmail        string `json:"ownerEmail"`
	WarrantyCode      string `json:"warrantyCode" binding:"required"`
	PolicyName        string `json:"policyName" binding:"required"`
	PolicyDescription string `json:"policyDescription"`
	DurationMonths    int    `json:"durationMonths" binding:"required"`
	Status            string `json:"status"`
	StartDate         string `json:"startDate"` // YYYY-MM-DD
	EndDate           string `json:"endDate"`   // YYYY-MM-DD
	InvoiceNumber     string `json:"invoiceNumber"`
	Note              string `json:"note"`
}

type UpdateWarrantyStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Note   string `json:"note"`
}

type CustomerRequestWarrantyRequest struct {
	SerialNumber string `json:"serialNumber" binding:"required"`
	ItemCode     string `json:"itemCode" binding:"required"`
	OwnerName    string `json:"ownerName" binding:"required"`
	OwnerEmail   string `json:"ownerEmail"`
	Note         string `json:"note"`
}

type ApproveWarrantyRequest struct {
	DurationMonths int    `json:"durationMonths" binding:"required"`
	PolicyName     string `json:"policyName" binding:"required"`
}

type RejectWarrantyRequest struct {
	Reason string `json:"reason" binding:"required"`
}
