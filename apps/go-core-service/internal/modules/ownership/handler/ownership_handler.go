package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/service"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type OwnershipHandler struct {
	ownershipService service.IOwnershipService
}

func NewOwnershipHandler(ownershipService service.IOwnershipService) *OwnershipHandler {
	return &OwnershipHandler{ownershipService: ownershipService}
}

// HELPERS 

func getUserID(c *gin.Context) (uuid.UUID, bool) {
	raw := c.GetHeader("X-User-Id")
	id, err := uuid.Parse(raw)
	if err != nil {
		apperror.HandleError(c, apperror.NewUnauthorized("Login required"))
		return uuid.Nil, false
	}
	return id, true
}

func getRole(c *gin.Context) string {
	return c.GetHeader("X-User-Role") // e.g. "ADMIN" or "CUSTOMER"
}

// REQUEST OTP
// POST /api/ownership/request-otp
// Cùng 1 endpoint, nhưng router tách ra 2 route khác nhau theo role.

func (h *OwnershipHandler) CustomerRequestOTP(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req dto.CustomerRequestOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	if err := h.ownershipService.CustomerRequestOTP(c.Request.Context(), req, userID); err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("OTP đã được gửi về email của bạn", nil))
}

func (h *OwnershipHandler) AdminRequestOTP(c *gin.Context) {
	adminID, ok := getUserID(c)
	if !ok {
		return
	}

	var req dto.AdminRequestOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	if err := h.ownershipService.AdminRequestOTP(c.Request.Context(), req, adminID); err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("OTP đã được gửi về email khách hàng", nil))
}

// VERIFY & REGISTER
// POST /api/ownership/register

func (h *OwnershipHandler) CustomerVerifyAndRegister(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	var req dto.CustomerVerifyAndRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	res, err := h.ownershipService.CustomerVerifyAndRegister(c.Request.Context(), req, userID)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Đăng ký quyền sở hữu thành công", res))
}

func (h *OwnershipHandler) AdminVerifyAndRegister(c *gin.Context) {
	adminID, ok := getUserID(c)
	if !ok {
		return
	}

	var req dto.AdminVerifyAndRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	res, err := h.ownershipService.AdminVerifyAndRegister(c.Request.Context(), req, adminID)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Đã đăng ký quyền sở hữu cho khách hàng thành công", res))
}

// GET OWNERSHIP DETAIL
// GET /api/ownership/:product_item_id (All authenticated roles)

func (h *OwnershipHandler) GetOwnershipDetail(c *gin.Context) {
	rawID := c.Param("product_item_id")
	productItemID, err := uuid.Parse(rawID)
	if err != nil {
		apperror.HandleError(c, apperror.NewValidation("product_item_id không hợp lệ"))
		return
	}

	res, err := h.ownershipService.GetOwnershipDetail(c.Request.Context(), productItemID)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Thông tin sở hữu", res))
}

