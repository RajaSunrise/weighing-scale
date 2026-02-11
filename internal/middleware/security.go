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

		// Defense in Depth: Force HTTPS
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Controls how much referrer information is included with requests
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (CSP)
		/*
		  SECURITY: CSP Hardening & Fix
		  - Risk: XSS and data exfiltration.
		  - Scenario: Injected scripts could steal sessions or deface the site.
		  - Mitigation: Tightened img-src and added missing domains for Hero image.
		  Note: 'unsafe-inline' is retained for compatibility with existing UI logic but should be moved to nonces in future.
		*/
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' cdn.tailwindcss.com cdn.jsdelivr.net unpkg.com; style-src 'self' 'unsafe-inline' fonts.googleapis.com; font-src 'self' fonts.gstatic.com; img-src 'self' data: https://www.transparenttextures.com https://lh3.googleusercontent.com; connect-src 'self';")

		c.Next()
	}
}
