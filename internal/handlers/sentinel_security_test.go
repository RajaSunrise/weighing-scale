package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"stoneweigh/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeadersAndCSP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecurityHeaders())
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify CSP
	csp := w.Header().Get("Content-Security-Policy")
	assert.Contains(t, csp, "https://lh3.googleusercontent.com")
	assert.Contains(t, csp, "default-src 'self'")

	// Verify HSTS
	hsts := w.Header().Get("Strict-Transport-Security")
	assert.Contains(t, hsts, "max-age=31536000")

	// Verify removal of deprecated X-XSS-Protection
	xss := w.Header().Get("X-XSS-Protection")
	assert.Empty(t, xss, "X-XSS-Protection should be removed as it is deprecated")
}

func TestTriggerANPR_Concurrency(t *testing.T) {
	r, server := setupTestServer(t)
	r.GET("/api/anpr/trigger", server.TriggerANPR)

	// We need to simulate a long running ANPR or just fill the semaphore
	// Since anprSemaphore is a package variable, we can access it if we are in the same package

	// Fill the semaphore (limit is 3)
	for i := 0; i < 3; i++ {
		anprSemaphore <- struct{}{}
	}
	defer func() {
		// Clean up semaphore for other tests
		for i := 0; i < 3; i++ {
			<-anprSemaphore
		}
	}()

	// Try one more request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/anpr/trigger?scale_id=1", nil)
	// Mock session
	r.POST("/login_test", func(c *gin.Context) {
		// Just to set session
		c.Status(200)
	})

	// Skip actual ANPR logic by testing the semaphore first
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
