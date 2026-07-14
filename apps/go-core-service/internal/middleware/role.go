package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	UserEntity "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/entity"
	Response "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(allowedRoles) == 0 {
			c.Next()
			return
		}

		var userRole string

		// 1. Try to get role directly from context (populated by AuthMiddleware)
		if r, exists := c.Get("role"); exists {
			if rStr, ok := r.(string); ok {
				userRole = rStr
			} else if rRole, ok := r.(UserEntity.Role); ok {
				userRole = string(rRole)
			}
		}

		// 2. Fallback to check current_user in context if role is not set directly
		if userRole == "" {
			if value, exists := c.Get("current_user"); exists {
				if user, ok := value.(*UserEntity.User); ok {
					userRole = string(user.Role)
				}
			}
		}

		if userRole == "" {
			c.JSON(http.StatusUnauthorized, Response.ResponseError(
				"unauthorized",
				"user role not found in context",
			))
			c.Abort()
			return
		}

		// Check if user's role is in the allowed list
		isAllowed := false
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, Response.ResponseError(
				"forbidden",
				"access denied: insufficient permissions",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}
