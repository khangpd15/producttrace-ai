package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS) headers.
// It reads the allowed origins from the FRONTEND_URL environment variable
// and always allows http://localhost:5173 for local development.
//
// Strategy: Go handles CORS (not Kong) — do NOT enable Kong CORS plugin simultaneously.
func CORSMiddleware() gin.HandlerFunc {
	// Build allowed origins list at startup (not per-request) for performance.
	allowedOrigins := buildAllowedOrigins()

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Only set CORS headers when an Origin header is present
		// (i.e. this is a cross-origin request from a browser).
		if origin != "" && isAllowedOrigin(origin, allowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept")
			c.Header("Vary", "Origin")
		}

		// Handle preflight OPTIONS — must return 200 immediately without
		// going through JWT/Auth middleware.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// buildAllowedOrigins reads FRONTEND_URL from the environment and
// always adds the local dev origin.
func buildAllowedOrigins() []string {
	origins := []string{
		"http://localhost:5173",
		"http://localhost:3000",
		"http://localhost:3001",
	}

	if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
		// Support comma-separated list of URLs: FRONTEND_URL=https://a.vercel.app,https://b.vercel.app
		for _, u := range strings.Split(frontendURL, ",") {
			u = strings.TrimSpace(u)
			if u != "" {
				origins = append(origins, u)
			}
		}
	}

	return origins
}

// isAllowedOrigin checks whether the given origin is in the allowed list.
func isAllowedOrigin(origin string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(a, origin) {
			return true
		}
	}
	return false
}
