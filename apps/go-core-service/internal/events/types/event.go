package types

import "time"

type Event struct {
	EventID       string      `json:"eventId"`
	EventType     string      `json:"eventType"`
	EventVersion  string      `json:"eventVersion"`
	Timestamp     time.Time   `json:"timestamp"`
	Producer      string      `json:"producer"`
	CorrelationID string      `json:"correlationId"`
	Payload       interface{} `json:"payload"`
}
