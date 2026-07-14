package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/location/service"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type LocationHandler struct {
	svc service.LocationService
}

func NewLocationHandler(svc service.LocationService) *LocationHandler {
	return &LocationHandler{svc: svc}
}

func (h *LocationHandler) Create(c *gin.Context) {
	var req dto.CreateLocationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		if strings.Contains(err.Error(), "openingHoursJson") || strings.Contains(err.Error(), "OpeningHoursJSON") {
			apperror.HandleError(c, &apperror.AppError{
				Code:    "INVALID_OPENING_HOURS",
				Message: "Định dạng giờ mở cửa không hợp lệ.",
				Status:  http.StatusBadRequest,
			})
			return
		}
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	loc, err := h.svc.CreateLocation(c.Request.Context(), &req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.ResponseSuccess("Location created successfully", loc))
}

func (h *LocationHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		apperror.HandleError(c, apperror.NewBadRequest("id is required"))
		return
	}

	loc, err := h.svc.GetLocationByID(c.Request.Context(), id)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("OK", loc))
}

// GetAll handles GET /locations
func (h *LocationHandler) GetAll(c *gin.Context) {
	var req dto.ListLocationsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	result, err := h.svc.ListLocations(c.Request.Context(), &req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("OK", result))
}

func (h *LocationHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		apperror.HandleError(c, apperror.NewBadRequest("id is required"))
		return
	}

	var req dto.UpdateLocationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		if strings.Contains(err.Error(), "openingHoursJson") || strings.Contains(err.Error(), "OpeningHoursJSON") {
			apperror.HandleError(c, &apperror.AppError{
				Code:    "INVALID_OPENING_HOURS",
				Message: "Định dạng giờ mở cửa không hợp lệ.",
				Status:  http.StatusBadRequest,
			})
			return
		}
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	loc, err := h.svc.UpdateLocation(c.Request.Context(), id, &req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Location updated successfully", loc))
}

func (h *LocationHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		apperror.HandleError(c, apperror.NewBadRequest("id is required"))
		return
	}

	if err := h.svc.HardDeleteLocation(c.Request.Context(), id); err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Location deleted successfully", nil))
}
