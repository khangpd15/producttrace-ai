package utils

import (
	"github.com/gin-gonic/gin"
)

// GetCurrentUserID retrieves the authenticated user's ID (UUID string) from the context.
func GetCurrentUserID(c *gin.Context) string {
	if val, exists := c.Get("user_id"); exists {
		if idStr, ok := val.(string); ok {
			return idStr
		}
	}
	return ""
}

// GetCurrentRole retrieves the authenticated user's role from the context.
func GetCurrentRole(c *gin.Context) string {
	if val, exists := c.Get("role"); exists {
		if roleStr, ok := val.(string); ok {
			return roleStr
		}
	}
	return ""
}

// GetCurrentEmail retrieves the authenticated user's email from the context.
func GetCurrentEmail(c *gin.Context) string {
	if val, exists := c.Get("email"); exists {
		if emailStr, ok := val.(string); ok {
			return emailStr
		}
	}
	return ""
}
