package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a Gin middleware that sets permissive CORS headers.
// allowedOrigins is the list of permitted origins; pass ["*"] to allow all.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[strings.ToLower(o)] = struct{}{}
	}
	allowAll := len(allowedOrigins) == 0 || allowedOrigins[0] == "*"

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		allowed := false
		if allowAll {
			allowed = true
		} else if _, ok := originSet[strings.ToLower(origin)]; ok {
			allowed = true
		}

		if allowed {
			if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			} else {
				c.Header("Access-Control-Allow-Origin", "*")
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
