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
	if val, exists := c.Get("user_id"); exists {
		if uid, ok := val.(uuid.UUID); ok {
			return uid, true
		}
		if str, ok := val.(string); ok {
			if uid, err := uuid.Parse(str); err == nil {
				return uid, true
			}
		}
	}
	raw := c.GetHeader("X-User-Id")
	id, err := uuid.Parse(raw)
	if err != nil {
		apperror.HandleError(c, apperror.NewUnauthorized("Login required"))
		return uuid.Nil, false
	}
	return id, true
}

func getRole(c *gin.Context) string {
	if val, exists := c.Get("role"); exists {
		if roleStr, ok := val.(string); ok {
			return roleStr
		}
	}
	return c.GetHeader("X-User-Role")
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

// ---------------------------------------------------------------------------
// CRUD Extensions
// ---------------------------------------------------------------------------

func (h *OwnershipHandler) TransferOwnership(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewValidation("ID không hợp lệ"))
		return
	}

	var req dto.TransferOwnershipReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		return // error is handled inside getUserId
	}
	role := getRole(c)

	if err := h.ownershipService.TransferOwnership(c.Request.Context(), id, req, userID, role); err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Chuyển quyền sở hữu thành công", nil))
}

func (h *OwnershipHandler) DeleteOwnership(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewValidation("ID không hợp lệ"))
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		return
	}
	role := getRole(c)

	if err := h.ownershipService.DeleteOwnership(c.Request.Context(), id, userID, role); err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Xóa quyền sở hữu thành công", nil))
}

func (h *OwnershipHandler) SearchOwnerships(c *gin.Context) {
	var req dto.SearchOwnershipsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	userID, ok := getUserID(c)
	if !ok {
		return
	}
	role := getRole(c)

	res, err := h.ownershipService.SearchOwnerships(c.Request.Context(), req, userID, role)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Danh sách quyền sở hữu", res))
}

