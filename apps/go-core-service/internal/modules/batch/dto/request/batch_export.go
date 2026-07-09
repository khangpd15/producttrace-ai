package request

type ExportBatchRequest struct {
	DestinationLocation string `json:"destination_location" binding:"required"`
	Quantity            int    `json:"quantity" binding:"required,min=1"`
	OperatorName        string `json:"operator_name" binding:"required"`
	Notes               string `json:"notes"`
}
