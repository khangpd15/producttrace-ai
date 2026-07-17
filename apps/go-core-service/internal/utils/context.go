package utils

import (
	"context"

	"github.com/gin-gonic/gin"
)

type contextKey string

const ActorIDKey contextKey = "actor_id"

// WithActorID returns a new context with the actor ID set.
func WithActorID(ctx context.Context, actorID string) context.Context {
	return context.WithValue(ctx, ActorIDKey, actorID)
}

// GetActorID retrieves the actor ID from the context.
func GetActorID(ctx context.Context) string {
	if val := ctx.Value(ActorIDKey); val != nil {
		if idStr, ok := val.(string); ok {
			return idStr
		}
	}
	return ""
}

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
