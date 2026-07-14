package request

import "time"

type GetAuditLogsRequest struct {
	Page     int        `form:"page" binding:"omitempty,min=1"`
	Limit    int        `form:"limit" binding:"omitempty,min=1,max=100"`
	Action   string     `form:"action" binding:"omitempty,oneof=CREATE UPDATE DELETE"`
	Entity   string     `form:"entity" binding:"omitempty"`
	FromDate *time.Time `form:"from_date" time_format:"2006-01-02T15:04:05Z07:00"`
	ToDate   *time.Time `form:"to_date" time_format:"2006-01-02T15:04:05Z07:00"`
}
