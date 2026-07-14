package dto

import "time"

type DashboardChartItem struct {
	TimePeriod       time.Time `json:"time_period"`
	ProductionVolume float64   `json:"production_volume"`
	SalesVolume      float64   `json:"sales_volume"`
}
