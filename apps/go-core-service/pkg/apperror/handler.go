package apperror

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/khangpd15/producttrace-ai/apps/go-core-service/pkg/response"
)

// HandleError inspects the error and writes the appropriate JSON response.
// If the error is an *AppError, it uses the embedded HTTP status and error code.
// Otherwise it falls back to 500 Internal Server Error.
func HandleError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.Status, response.ResponseError(appErr.Message, appErr.Code))
		return
	}
	// Unknown / non-AppError → 500
	c.JSON(500, response.ResponseError("Internal Server Error", CodeInternal))
}
