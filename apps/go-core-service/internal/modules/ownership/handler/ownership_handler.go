package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/dto"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/ownership/service"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type SSEBroker struct {
	AdminClients    map[chan string]bool
	CustomerClients map[string]map[chan string]bool // map[userID] -> map[chan]bool
}

type OwnershipHandler struct {
	ownershipService service.IOwnershipService
	broker           *SSEBroker
}

func NewOwnershipHandler(ownershipService service.IOwnershipService) *OwnershipHandler {
	return &OwnershipHandler{
		ownershipService: ownershipService,
		broker: &SSEBroker{
			AdminClients:    make(map[chan string]bool),
			CustomerClients: make(map[string]map[chan string]bool),
		},
	}
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

	productID, err := h.ownershipService.CustomerRequestOTP(c.Request.Context(), req, userID)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("OTP đã được gửi về email của bạn", gin.H{"product_id": productID}))
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

	productID, err := h.ownershipService.AdminRequestOTP(c.Request.Context(), req, adminID)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("OTP đã được gửi về email khách hàng", gin.H{"product_id": productID}))
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

	// Broadcast to active Admin SSE clients
	for client := range h.broker.AdminClients {
		select {
		case client <- "NEW_OWNERSHIP_REQUEST":
		default:
			// Client channel is full or blocked, skip to prevent hanging handler
		}
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

// ---------------------------------------------------------------------------
// ADMIN APPROVAL & SSE ENDPOINTS
// ---------------------------------------------------------------------------

type ApproveReq struct {
	OwnershipID uuid.UUID `json:"ownership_id" binding:"required"`
}

func (h *OwnershipHandler) ApproveOwnership(c *gin.Context) {
	adminID, ok := getUserID(c)
	if !ok { return }

	var req ApproveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	if err := h.ownershipService.ApproveOwnership(c.Request.Context(), req.OwnershipID, adminID); err != nil {
		apperror.HandleError(c, err)
		return
	}

	// Try to find the ownership to notify the correct customer
    // We would need the customer ID to target the exact channel, but it's simpler to broadcast 
    // to all customers and let FE filter, since it's just a demo.
    for _, userChans := range h.broker.CustomerClients {
        for client := range userChans {
			select {
			case client <- "OWNERSHIP_VERDICT":
			default:
			}
        }
    }

	c.JSON(200, response.ResponseSuccess("Đã duyệt đăng ký sở hữu thành công", nil))
}

func (h *OwnershipHandler) RejectOwnership(c *gin.Context) {
	adminID, ok := getUserID(c)
	if !ok { return }

	var req ApproveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	if err := h.ownershipService.RejectOwnership(c.Request.Context(), req.OwnershipID, adminID); err != nil {
		apperror.HandleError(c, err)
		return
	}

    for _, userChans := range h.broker.CustomerClients {
        for client := range userChans {
			select {
			case client <- "OWNERSHIP_VERDICT":
			default:
			}
        }
    }

	c.JSON(200, response.ResponseSuccess("Đã từ chối đăng ký sở hữu", nil))
}

func (h *OwnershipHandler) AdminStream(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	messageChan := make(chan string, 10)
	h.broker.AdminClients[messageChan] = true

	defer func() {
		delete(h.broker.AdminClients, messageChan)
		close(messageChan)
	}()

	clientGone := c.Writer.CloseNotify()
	for {
		select {
		case <-clientGone:
			return
		case msg := <-messageChan:
			c.Writer.Write([]byte("data: " + msg + "\n\n"))
			c.Writer.Flush()
		}
	}
}

func (h *OwnershipHandler) CustomerStream(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	userID, ok := getUserID(c)
	if !ok { return }

	messageChan := make(chan string, 10)
	if h.broker.CustomerClients[userID.String()] == nil {
		h.broker.CustomerClients[userID.String()] = make(map[chan string]bool)
	}
	h.broker.CustomerClients[userID.String()][messageChan] = true

	defer func() {
		delete(h.broker.CustomerClients[userID.String()], messageChan)
		close(messageChan)
	}()

	clientGone := c.Writer.CloseNotify()
	for {
		select {
		case <-clientGone:
			return
		case msg := <-messageChan:
			c.Writer.Write([]byte("data: " + msg + "\n\n"))
			c.Writer.Flush()
		}
	}
}

