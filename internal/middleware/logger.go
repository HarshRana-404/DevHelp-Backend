package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger returns a Gin middleware that logs each request using the provided Zap logger.
// The following fields are logged per request:
//   - timestamp, request_id, ip, method, path, status, latency,
//     user_agent, request_size, response_size
func Logger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		requestID, _ := c.Get(requestIDKey)

		fields := []zap.Field{
			zap.String("request_id", toString(requestID)),
			zap.String("ip", c.ClientIP()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int64("request_size", c.Request.ContentLength),
			zap.Int("response_size", c.Writer.Size()),
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Error("request error", append(fields, zap.Error(e))...)
			}
		} else {
			switch {
			case status >= 500:
				log.Error("request", fields...)
			case status >= 400:
				log.Warn("request", fields...)
			default:
				log.Info("request", fields...)
			}
		}
	}
}

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
