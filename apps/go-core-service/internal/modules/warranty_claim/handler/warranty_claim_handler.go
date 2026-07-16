package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/warranty_claim/service"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type WarrantyClaimHandler struct {
	service service.WarrantyClaimService
}

func NewWarrantyClaimHandler(s service.WarrantyClaimService) *WarrantyClaimHandler {
	return &WarrantyClaimHandler{service: s}
}

func toAppError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, service.ErrNoActiveWarranty) {
		return apperror.NewBadRequest("Không tìm thấy hợp đồng bảo hành hợp lệ cho thiết bị này.")
	}
	if errors.Is(err, service.ErrClaimNotFound) {
		return apperror.NewNotFound("Không tìm thấy yêu cầu bảo hành.")
	}
	if errors.Is(err, service.ErrInvalidStatus) {
		return apperror.NewBadRequest("Trạng thái yêu cầu bảo hành không hợp lệ.")
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

func (h *WarrantyClaimHandler) CreateClaim(c *gin.Context) {
	var req dto.CreateClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	resp, err := h.service.CustomerCreateClaim(c.Request.Context(), req)
	if err != nil {
		apperror.HandleError(c, toAppError(err))
		return
	}

	c.JSON(201, response.ResponseSuccess("Gửi yêu cầu bảo hành thành công", resp))
}

func (h *WarrantyClaimHandler) UpdateClaimStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewValidation("ID yêu cầu bảo hành không hợp lệ"))
		return
	}

	var req dto.UpdateClaimStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	resp, err := h.service.AdminUpdateClaimStatus(c.Request.Context(), id, req)
	if err != nil {
		apperror.HandleError(c, toAppError(err))
		return
	}

	c.JSON(200, response.ResponseSuccess("Cập nhật trạng thái yêu cầu bảo hành thành công", resp))
}

func (h *WarrantyClaimHandler) ListAllClaims(c *gin.Context) {
	resp, err := h.service.ListAllClaims(c.Request.Context())
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Lấy danh sách yêu cầu bảo hành thành công", resp))
}

func (h *WarrantyClaimHandler) ListMyClaims(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		apperror.HandleError(c, apperror.NewUnauthorized(err.Error()))
		return
	}

	resp, err := h.service.ListMyClaims(c.Request.Context(), userID)
	if err != nil {
		apperror.HandleError(c, toAppError(err))
		return
	}

	c.JSON(200, response.ResponseSuccess("Lấy danh sách yêu cầu bảo hành của tôi thành công", resp))
}

func (h *WarrantyClaimHandler) GetClaimByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apperror.HandleError(c, apperror.NewValidation("ID yêu cầu bảo hành không hợp lệ"))
		return
	}

	resp, err := h.service.GetClaimByID(c.Request.Context(), id)
	if err != nil {
		apperror.HandleError(c, toAppError(err))
		return
	}

	c.JSON(200, response.ResponseSuccess("Lấy chi tiết yêu cầu bảo hành thành công", resp))
}
