package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/dashboard/services"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type DashboardHandler struct {
	dashboardService services.DashboardService
}

func NewDashboardHandler(dashboardService services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

func handleError(c *gin.Context, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus(), response.ResponseError(appErr.Message, nil))
		return
	}
	c.JSON(http.StatusInternalServerError, response.ResponseError("Internal server error", nil))
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	stats, err := h.dashboardService.GetStats(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Get dashboard stats successfully", stats))
}

func (h *DashboardHandler) GetActivities(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	activities, err := h.dashboardService.GetActivities(c.Request.Context(), limit)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Get dashboard activities successfully", activities))
}

func (h *DashboardHandler) GetAlerts(c *gin.Context) {
	alerts, err := h.dashboardService.GetAlerts(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Get dashboard alerts successfully", alerts))
}

func (h *DashboardHandler) GetProductionSalesChart(c *gin.Context) {
	chartData, err := h.dashboardService.GetProductionSalesChart(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Get production and sales chart data successfully", chartData))
}
