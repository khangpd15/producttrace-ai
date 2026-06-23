package handler

import (
	"net/http"

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
		c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
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
		c.JSON(http.StatusBadRequest, response.ResponseError("id is required", nil))
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
		c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
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
		c.JSON(http.StatusBadRequest, response.ResponseError("id is required", nil))
		return
	}

	var req dto.UpdateLocationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
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
		c.JSON(http.StatusBadRequest, response.ResponseError("id is required", nil))
		return
	}

	if err := h.svc.DeleteLocation(c.Request.Context(), id); err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Location deleted successfully", nil))
}

func (h *LocationHandler) FindNearby(c *gin.Context) {
	var req dto.FindNearbyReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ResponseError(err.Error(), nil))
		return
	}

	locations, err := h.svc.FindNearby(c.Request.Context(), &req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("OK", locations))
}
