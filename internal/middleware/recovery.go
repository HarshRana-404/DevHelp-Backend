package middleware

import (
	"net/http"

	"devhelp/internal/services/dto"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery returns a Gin middleware that recovers from panics, logs the panic details,
// and returns a 500 JSON response instead of crashing the server.
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get(requestIDKey)
				log.Error("panic recovered",
					zap.String("request_id", toString(requestID)),
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.APIResponse{
					Success: false,
					Error: &dto.APIError{
						Code:    "INTERNAL_ERROR",
						Message: "An unexpected error occurred",
					},
				})
			}
		}()
		c.Next()
	}
}
