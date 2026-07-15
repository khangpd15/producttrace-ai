package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/service"
)

type WarrantyHandler struct {
	service service.WarrantyService
}

func NewWarrantyHandler(s service.WarrantyService) *WarrantyHandler {
	return &WarrantyHandler{service: s}
}

func (h *WarrantyHandler) ActivateWarranty(c *gin.Context) {
	var req dto.CreateWarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format or missing required fields", "details": err.Error()})
		return
	}

	resp, err := h.service.ActivateWarranty(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrWarrantyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to activate warranty", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "warranty activated successfully",
		"data":    resp,
	})
}

func (h *WarrantyHandler) RequestWarranty(c *gin.Context) {
	var req dto.CustomerRequestWarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format", "details": err.Error()})
		return
	}

	resp, err := h.service.RequestWarranty(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrWarrantyExists) || errors.Is(err, service.ErrNotOwned) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to request warranty", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "warranty requested successfully",
		"data":    resp,
	})
}

func (h *WarrantyHandler) ApproveWarranty(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warranty id"})
		return
	}

	var req dto.ApproveWarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format", "details": err.Error()})
		return
	}

	resp, err := h.service.ApproveWarranty(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrWarrantyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrOver30Days) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve warranty", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "warranty approved",
		"data":    resp,
	})
}

func (h *WarrantyHandler) RejectWarranty(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid warranty id"})
		return
	}

	var req dto.RejectWarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format", "details": err.Error()})
		return
	}

	resp, err := h.service.RejectWarranty(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrWarrantyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject warranty", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "warranty rejected",
		"data":    resp,
	})
}

func (h *WarrantyHandler) ListWarranties(c *gin.Context) {
	resp, err := h.service.ListWarranties(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list warranties", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}
