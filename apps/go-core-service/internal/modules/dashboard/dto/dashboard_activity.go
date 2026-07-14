package dto

import "time"

type DashboardActivity struct {
	ID          string    `json:"id"`
	EventType   string    `json:"event_type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
