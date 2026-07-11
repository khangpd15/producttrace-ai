package dto

type DashboardStats struct {
	TotalProducts        int64 `json:"total_products"`
	TotalBatches         int64 `json:"total_batches"`
	TotalOwnerships      int64 `json:"total_ownerships"`
	TotalUnderWarranty   int64 `json:"total_under_warranty"`
	TotalPendingApproval int64 `json:"total_pending_approval"`
	TotalLocations       int64 `json:"total_locations"`
}
