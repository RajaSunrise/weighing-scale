package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds security-related headers to the response
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Protects against MIME sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Protects against clickjacking (allows framing only by same origin)
		c.Header("X-Frame-Options", "SAMEORIGIN")

		// Enables XSS filtering in browsers (deprecated in modern browsers but good for legacy)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Controls how much referrer information is included with requests
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}
