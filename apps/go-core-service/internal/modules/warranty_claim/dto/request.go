package dto

type CreateClaimRequest struct {
	SerialNumber     string `json:"serialNumber" binding:"required"`
	IssueTitle       string `json:"issueTitle" binding:"required"`
	IssueDescription string `json:"issueDescription" binding:"required"`
	ContactPhone     string `json:"contactPhone" binding:"required"`
	ContactEmail     string `json:"contactEmail"`
}

type UpdateClaimStatusRequest struct {
	Status         string `json:"status" binding:"required"`
	ResolutionNote string `json:"resolutionNote"`
}
