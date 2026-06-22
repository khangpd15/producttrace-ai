package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/dto/request"
	authResponse "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/dto/response"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/service"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/apperror"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

type AuthenHandler struct {
	authService service.AuthenServiceInterface
}

func NewAuthenHandler(authService service.AuthenServiceInterface) *AuthenHandler {
	return &AuthenHandler{
		authService: authService,
	}
}

func (h *AuthenHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	user, err := h.authService.RegisterUser(c.Request.Context(), req.Email, req.Phone, req.FullName, req.Password)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(201, response.ResponseSuccess("User registered successfully. Please verify your email with OTP.", authResponse.RegisterResponse{
		FullName: user.FullName,
		Phone:    user.Phone,
		Email:    user.Email,
		Status:   string(user.Status),
	}))
}

func (h *AuthenHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	accessToken, refreshToken, err := h.authService.LoginUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Logged in successfully", authResponse.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}))
}

func (h *AuthenHandler) VerifyOTP(c *gin.Context) {
	var req request.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	err := h.authService.VerifyOTP(c.Request.Context(), req.Email, req.OTP)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Account verified successfully. You can now log in.", nil))
}

func (h *AuthenHandler) RefreshToken(c *gin.Context) {
	var req request.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.HandleError(c, apperror.NewValidation(err.Error()))
		return
	}

	accessToken, refreshToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		apperror.HandleError(c, err)
		return
	}

	c.JSON(200, response.ResponseSuccess("Token refreshed successfully", authResponse.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}))
}
