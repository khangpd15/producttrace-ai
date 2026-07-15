package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/service"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
)

type WarrantyHandler struct {
	warrantyService service.WarrantyService
}

func NewWarrantyHandler(warrantyService service.WarrantyService) *WarrantyHandler {
	return &WarrantyHandler{
		warrantyService: warrantyService,
	}
}

func (h *WarrantyHandler) RequestActivation(c *gin.Context) {
	var req dto.RequestWarrantyActivationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_FAILED"})
		return
	}

	userID := h.getUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID is required", "code": "ACCESS_DENIED"})
		return
	}

	resp, err := h.warrantyService.RequestActivation(c.Request.Context(), userID, req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Warranty activation requested successfully",
		"data":    resp,
	})
}

func (h *WarrantyHandler) ConfirmActivation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warranty ID format", "code": "VALIDATION_FAILED"})
		return
	}

	var req dto.ConfirmWarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_FAILED"})
		return
	}

	userID := h.getUserID(c)
	if userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID is required", "code": "ACCESS_DENIED"})
		return
	}

	resp, err := h.warrantyService.ConfirmActivation(c.Request.Context(), userID, id, req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Warranty confirmation processed",
		"data":    resp,
	})
}

func (h *WarrantyHandler) GetWarrantyByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warranty ID format", "code": "VALIDATION_FAILED"})
		return
	}

	resp, err := h.warrantyService.GetWarrantyByID(c.Request.Context(), id)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": resp,
	})
}

func (h *WarrantyHandler) ListWarranties(c *gin.Context) {
	search := c.Query("search")
	status := c.Query("status")
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Role check: If customer, only show their own warranties. If admin/staff/dealer, show all or filtered.
	userID := h.getUserID(c)
	role := h.getUserRole(c)

	var ownerID *uuid.UUID
	if role == "CUSTOMER" && userID != uuid.Nil {
		ownerID = &userID
	}

	resp, err := h.warrantyService.ListWarranties(c.Request.Context(), search, status, ownerID, page, limit)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": resp.Items,
		"meta": gin.H{
			"total": resp.Total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *WarrantyHandler) UpdateWarranty(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warranty ID format", "code": "VALIDATION_FAILED"})
		return
	}

	var req dto.UpdateWarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_FAILED"})
		return
	}

	userID := h.getUserID(c)
	resp, err := h.warrantyService.UpdateWarranty(c.Request.Context(), userID, id, req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Warranty updated successfully",
		"data":    resp,
	})
}

func (h *WarrantyHandler) DeleteWarranty(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid warranty ID format", "code": "VALIDATION_FAILED"})
		return
	}

	userID := h.getUserID(c)
	err = h.warrantyService.DeleteWarranty(c.Request.Context(), userID, id)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Warranty deleted successfully",
	})
}

func (h *WarrantyHandler) GetStats(c *gin.Context) {
	stats, err := h.warrantyService.GetStats(c.Request.Context())
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

func (h *WarrantyHandler) getUserID(c *gin.Context) uuid.UUID {
	userIDVal, exists := c.Get("user_id")
	if exists {
		if idStr, ok := userIDVal.(string); ok {
			id, err := uuid.Parse(idStr)
			if err == nil {
				return id
			}
		} else if uid, ok := userIDVal.(uuid.UUID); ok {
			return uid
		}
	}

	userIDStr := c.GetHeader("X-User-Id")
	if userIDStr == "" {
		userIDStr = c.GetHeader("X-User-ID")
	}
	if userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err == nil {
			return id
		}
	}
	return uuid.Nil
}

func (h *WarrantyHandler) getUserRole(c *gin.Context) string {
	roleVal, exists := c.Get("role")
	if exists {
		if roleStr, ok := roleVal.(string); ok {
			return roleStr
		}
	}
	return c.GetHeader("X-User-Role")
}
