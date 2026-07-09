package response

import "time"

type BatchEventDTO struct {
	EventName string    `json:"event_name"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}
