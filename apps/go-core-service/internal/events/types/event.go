package types

import "time"

type Event struct {
    EventID string `json:"event_id"`
    EventType string `json:"event_type"`
    EventVersion string `json:"event_version"`
	Timestamp     time.Time   `json:"timestamp"`
    Producer string `json:"producer"`
    CorrelationID string `json:"correlation_id"`
	Payload       interface{} `json:"payload"`
}
