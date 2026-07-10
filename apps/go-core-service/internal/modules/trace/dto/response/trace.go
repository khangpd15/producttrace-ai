package response

type ProductItemDTO struct {
	ItemID       string `json:"itemId"`
	ItemCode     string `json:"itemCode"`
	SerialNumber string `json:"serialNumber"`
	Status       string `json:"status"`
	ProductName  string `json:"productName"`
	ThumbnailURL string `json:"thumbnailUrl"`
}

type TimelineEventDTO struct {
	EventID     string `json:"eventId"`
	EventType   string `json:"eventType"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Timestamp   string `json:"timestamp"`
}

type FilterAppliedDTO struct {
	FromDate   *string  `json:"fromDate,omitempty"`
	ToDate     *string  `json:"toDate,omitempty"`
	EventTypes []string `json:"eventTypes,omitempty"`
}

type TraceSearchResponse struct {
	ProductItem        *ProductItemDTO   `json:"productItem,omitempty"`
	Warning            string            `json:"warning,omitempty"`
	FilterApplied      *FilterAppliedDTO `json:"filterApplied,omitempty"`
	MatchedEventsCount *int              `json:"matchedEventsCount,omitempty"`
	Timeline           []TimelineEventDTO `json:"timeline"`
}

type ExportJobResponse struct {
	JobID                string `json:"jobId"`
	Status               string `json:"status"`
	EstimatedTimeSeconds int    `json:"estimatedTimeSeconds"`
}
