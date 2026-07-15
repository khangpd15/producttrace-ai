package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty/service"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type WarrantyHandler struct {
	service service.WarrantyService
}

func NewWarrantyHandler(s service.WarrantyService) *WarrantyHandler {
	return &WarrantyHandler{service: s}
}

func toAppError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, service.ErrWarrantyExists) {
		return apperror.NewConflict("Bảo hành cho sản phẩm này đã tồn tại.")
	}
	if errors.Is(err, service.ErrWarrantyNotFound) {
		return apperror.NewNotFound("Không tìm thấy thông tin bảo hành.")
	}
	if errors.Is(err, service.ErrNotOwned) {
		return apperror.NewBadRequest("Sản phẩm chưa được đăng ký sở hữu hoặc quyền sở hữu không hoạt động.")
	}
	if errors.Is(err, service.ErrOver30Days) {
		return apperror.NewBadRequest("Yêu cầu đăng ký bảo hành vượt quá 30 ngày kể từ ngày kích hoạt quyền sở hữu.")
	}
	return err
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	val, exists := c.Get("userId")
	if !exists {
		return uuid.Nil, errors.New("unauthorized: missing userId in context")
	}
	idStr, ok := val.(string)
	if !ok {
		return uuid.Nil, errors.New("unauthorized: invalid userId type")
	}
	return uuid.Parse(idStr)
}

func (h *WarrantyHandler) ActivateWarranty(c *gin.Context) {
	var req dto.CreateWarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	resp, err := h.service.ActivateWarranty(c.Request.Context(), req)
	if err != nil {
		apperror.HandleError(c, toAppError(err))
		return
	}

	c.JSON(201, response.ResponseSuccess("Kích hoạt bảo hành thành công", resp))
}

func (h *WarrantyHandler) RequestWarranty(c *gin.Context) {
	var req dto.CustomerRequestWarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	resp, err := h.service.RequestWarranty(c.Request.Context(), req)
	if err != nil {
		apperror.HandleError(c, toAppError(err))
		return
	}

	c.JSON(201, response.ResponseSuccess("Gửi yêu cầu bảo hành thành công", resp))
}

func (h *WarrantyHandler) ApproveWarranty(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewValidation("ID bảo hành không hợp lệ"))
		return
	}

	var req dto.ApproveWarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	resp, err := h.service.ApproveWarranty(c.Request.Context(), id, req)
	if err != nil {
		apperror.HandleError(c, toAppError(err))
		return
	}

	c.JSON(200, response.ResponseSuccess("Duyệt bảo hành thành công", resp))
}

func (h *WarrantyHandler) RejectWarranty(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewValidation("ID bảo hành không hợp lệ"))
		return
	}

	var req dto.RejectWarrantyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	resp, err := h.service.RejectWarranty(c.Request.Context(), id, req)
	if err != nil {
		apperror.HandleError(c, toAppError(err))
		return
	}

	c.JSON(200, response.ResponseSuccess("Từ chối bảo hành thành công", resp))
}

func (h *WarrantyHandler) ListWarranties(c *gin.Context) {
	resp, err := h.service.ListWarranties(c.Request.Context())
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Lấy danh sách bảo hành thành công", resp))
}

func (h *WarrantyHandler) ListMyWarranties(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		apperror.HandleError(c, apperror.NewUnauthorized(err.Error()))
		return
	}

	resp, err := h.service.ListMyWarranties(c.Request.Context(), userID)
	if err != nil {
		apperror.HandleError(c, toAppError(err))
		return
	}

	c.JSON(200, response.ResponseSuccess("Lấy danh sách bảo hành của tôi thành công", resp))
}

func (h *WarrantyHandler) GetWarrantyByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewValidation("ID bảo hành không hợp lệ"))
		return
	}

	resp, err := h.service.GetWarrantyByID(c.Request.Context(), id)
	if err != nil {
		apperror.HandleError(c, toAppError(err))
		return
	}

	c.JSON(200, response.ResponseSuccess("Lấy chi tiết bảo hành thành công", resp))
}

func (h *WarrantyHandler) GetWarrantyByProductItemID(c *gin.Context) {
	idStr := c.Param("product_item_id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewValidation("ID sản phẩm không hợp lệ"))
		return
	}

	resp, err := h.service.GetWarrantyByProductItemID(c.Request.Context(), id)
	if err != nil {
		apperror.HandleError(c, toAppError(err))
		return
	}

	c.JSON(200, response.ResponseSuccess("Lấy thông tin bảo hành sản phẩm thành công", resp))
}

func (h *WarrantyHandler) VoidWarranty(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewValidation("ID bảo hành không hợp lệ"))
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	resp, err := h.service.VoidWarranty(c.Request.Context(), id, req.Reason)
	if err != nil {
		apperror.HandleError(c, toAppError(err))
		return
	}

	c.JSON(200, response.ResponseSuccess("Vô hiệu hóa bảo hành thành công", resp))
}
