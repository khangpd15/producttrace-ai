package response

// ExportBatchesResponse trả về kết quả sau khi xuất nhiều batch.
type ExportBatchesResponse struct {
	// ExportedBatchCount: số batch đã được xuất thành công.
	ExportedBatchCount int `json:"exported_batch_count"`
	// ExportedItemCount: tổng số ProductItem đã được xuất.
	ExportedItemCount int `json:"exported_item_count"`
	// BatchIDs: danh sách UUID các batch đã xuất.
	BatchIDs []string `json:"batch_ids"`
	// DestinationLocationID: UUID của location đích.
	DestinationLocationID string `json:"destination_location_id"`
}
