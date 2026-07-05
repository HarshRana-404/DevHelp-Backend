// Package middleware contains all Gin middleware components used by DevHelp.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDKey = "request_id"
const requestIDHeader = "X-Request-ID"

// RequestID injects a unique request identifier into every request context.
// If the client sends an X-Request-ID header its value is reused; otherwise a new UUID is generated.
// The ID is also echoed back in the response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(requestIDKey, id)
		c.Header(requestIDHeader, id)
		c.Next()
	}
}
