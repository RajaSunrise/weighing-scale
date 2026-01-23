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

		// Content Security Policy (CSP)
		// Allows scripts from Trusted CDNs (Tailwind, HTMX, AOS) and inline scripts (necessary for some UI logic)
		// Allows styles from Google Fonts
		// Allows images from transparenttextures.com (used in home hero)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' cdn.tailwindcss.com cdn.jsdelivr.net unpkg.com; style-src 'self' 'unsafe-inline' fonts.googleapis.com; font-src 'self' fonts.gstatic.com; img-src 'self' data: https://www.transparenttextures.com; connect-src 'self';")

		c.Next()
	}
}
