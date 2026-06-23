package middleware

import (
	"fmt"
	"log"
	"net/http"

	Response "github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var message string
				switch v := err.(type) {
				case error:
					message = v.Error()
				case string:
					message = v
				default:
					message = fmt.Sprintf("%v", v)
				}

				// Log the panic details for debugging
				log.Printf("[Panic Recovered] err: %s\n", message)

				c.JSON(http.StatusInternalServerError, Response.ResponseError(
					"Internal server error",
					message,
				))
				c.Abort()
			}
		}()

		c.Next()
	}
}