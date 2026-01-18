package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds standard security headers to all responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Protects against MIME sniffing vulnerabilities
		c.Header("X-Content-Type-Options", "nosniff")

		// Protects against Clickjacking attacks
		c.Header("X-Frame-Options", "SAMEORIGIN")

		// Enables the Cross-Site Scripting (XSS) filter built into most recent web browsers
		c.Header("X-XSS-Protection", "1; mode=block")

		// Controls how much referrer information is included with requests
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}
