package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestLogger_Sanitization(t *testing.T) {
	// Redirect log output to a buffer
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil) // Reset

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestLogger())
	r.GET("/test", func(c *gin.Context) {
		c.Status(200)
	})

	// Case 1: Malicious path with script tag
	maliciousPath := "/test/<script>alert(1)</script>"
	req, _ := http.NewRequest("GET", maliciousPath, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	logOutput := buf.String()
	assert.NotContains(t, logOutput, "<script>", "Log should not contain raw HTML tags")
	assert.Contains(t, logOutput, "script alert(1)", "Log should contain sanitized path content")
}
