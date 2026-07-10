package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	Repo "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/modules/user/repository"
	Utils "github.com/khangpd15/producttrace-ai/apps/go-core-service/internal/utils"
	Response "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

func AuthMiddleware(userRepo Repo.UserRepositoryInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, Response.ResponseError(
				"missing Authorization header",
				"authorization header is required",
			))
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, Response.ResponseError(
				"invalid Authorization format",
				"Authorization header must be in the format 'Bearer <token>'",
			))
			c.Abort()
			return
		}

		claims, err := Utils.ValidateAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, Response.ResponseError(
				"invalid or expired token",
				"token validation failed: "+err.Error(),
			))
			c.Abort()
			return
		}

		if claims.UserID == "" {
			c.JSON(http.StatusUnauthorized, Response.ResponseError(
				"invalid user ID",
				"invalid user ID in token",
			))
			c.Abort()
			return
		}

		// Store user details in context from JWT claims without querying the DB
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		
		// Set X-User-Id request header so downstream handlers (like GetProfile/UpdateProfile)
		// which retrieve it via c.GetHeader("X-User-Id") can access it.
		if c.Request != nil {
			if c.Request.Header == nil {
				c.Request.Header = make(http.Header)
			}
			c.Request.Header.Set("X-User-Id", claims.UserID)
		}

		c.Next()
	}
}
