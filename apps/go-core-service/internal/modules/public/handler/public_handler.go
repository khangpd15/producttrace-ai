package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/public/service"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type PublicHandler struct {
	svc service.PublicService
}

func NewPublicHandler(svc service.PublicService) *PublicHandler {
	return &PublicHandler{
		svc: svc,
	}
}

func (h *PublicHandler) VerifyQR(c *gin.Context) {
	itemCode := c.Query("item_code")
	token := c.Query("token")

	if itemCode == "" || token == "" {
		apperror.HandleError(c, apperror.NewBadRequest("item_code and token are required query parameters"))
		return
	}

	res, err := h.svc.VerifyQR(c.Request.Context(), itemCode, token)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.ResponseSuccess("Item verified successfully", res))
}
