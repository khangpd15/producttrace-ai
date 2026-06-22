package mapper

import (
	RequestAuthen "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/dto/request"
	ResponseAuthen "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/authen/dto/response"
	UserEntity "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/entity"
)

func AuthenToUserEntity(request *RequestAuthen.RegisterRequest, passwordHash string) *UserEntity.User {
	return &UserEntity.User{
		Email:        request.Email,
		Phone:        request.Phone,
		FullName:     request.FullName,
		PasswordHash: passwordHash,
		Role:         UserEntity.RoleCustomer,
		Status:       UserEntity.StatusActive,
	}
}
func UserEntityToAuthenResponse(user *UserEntity.User) *ResponseAuthen.RegisterResponse {
	return &ResponseAuthen.RegisterResponse{
		FullName: user.FullName,
		Phone:    user.Phone,
		Email:    user.Email,
		Status:   string(user.Status),
	}
}
func UserEntityToTokenResponse(accessToken string, refreshToken string) *ResponseAuthen.TokenResponse {
	return &ResponseAuthen.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}
