package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/service"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
)

type WarrantyClaimHandler struct {
	claimService service.WarrantyClaimService
}

func NewWarrantyClaimHandler(claimService service.WarrantyClaimService) *WarrantyClaimHandler {
	return &WarrantyClaimHandler{
		claimService: claimService,
	}
}

func (h *WarrantyClaimHandler) CreateWarrantyClaim(c *gin.Context) {
	var req dto.CreateWarrantyClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_FAILED"})
		return
	}

	// Note: in a real application the UserID would be extracted from the authentication token middleware. 
	// For example: userID := c.MustGet("user_id").(uuid.UUID)
	// We'll mock the extraction here for now.
	userIDStr := c.GetHeader("X-User-Id")
	if userIDStr == "" {
		userIDStr = c.GetHeader("X-User-ID")
	}
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID is required", "code": "ACCESS_DENIED"})
		return
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid User ID format", "code": "ACCESS_DENIED"})
		return
	}

	resp, err := h.claimService.CreateWarrantyClaim(c.Request.Context(), userID, req)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Warranty Claim Created",
		"data":    resp,
	})
}
