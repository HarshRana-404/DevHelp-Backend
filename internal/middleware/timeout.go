package middleware

import (
	"context"
	"net/http"
	"time"

	"devhelp/internal/services/dto"

	"github.com/gin-gonic/gin"
)

// Timeout returns a Gin middleware that cancels the request context after d duration.
// If the handler takes longer than d, a 504 Gateway Timeout response is returned.
//
// Implementation note: we do NOT spawn a goroutine for the handler itself because
// Gin's ResponseWriter is not goroutine-safe. Instead we rely on the context
// cancellation signal: handlers that respect ctx.Done() will abort early, and
// the middleware checks after c.Next() whether the deadline was exceeded.
// For truly long-running handlers, the HTTP server's WriteTimeout acts as the
// hard ceiling.
func Timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		// If the context deadline was exceeded and no response has been written yet,
		// return 504. Gin may have already written headers in a handler, in which
		// case we cannot change the status — so we only act when the writer is still
		// uncommitted.
		if ctx.Err() == context.DeadlineExceeded && !c.Writer.Written() {
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, dto.APIResponse{
				Success: false,
				Error: &dto.APIError{
					Code:    "TIMEOUT",
					Message: "Request timed out",
				},
			})
		}
	}
}
