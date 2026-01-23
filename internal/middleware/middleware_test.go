package middleware

import (
	"encoding/gob"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func init() {
	// Register type for gob encoding in session
	gob.Register(time.Time{})
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))
	return r
}

func TestSecurityHeaders(t *testing.T) {
	r := setupTestRouter()
	r.Use(SecurityHeaders())
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "SAMEORIGIN", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'self'")
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "cdn.tailwindcss.com")
}

func TestAuthRequired_NoSession(t *testing.T) {
	r := setupTestRouter()
	r.Use(AuthRequired())
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Case 1: HTML Request -> Redirect
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))

	// Case 2: JSON Request -> 401
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/protected", nil)
	req2.Header.Set("Accept", "application/json")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)
}

func TestAuthRequired_WithSession(t *testing.T) {
	r := setupTestRouter()

	// Helper to set session
	r.GET("/login", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("user_id", 1)
		session.Save()
		c.Status(http.StatusOK)
	})

	r.Use(AuthRequired())
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 1. Perform Login to get cookie
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/login", nil)
	r.ServeHTTP(w, req)

	// 2. Access protected route with cookie
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/protected", nil)
	req2.Header.Set("Cookie", w.Header().Get("Set-Cookie"))
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestRoleRequired(t *testing.T) {
	r := setupTestRouter()

	// Helper to set session
	r.GET("/login/:role", func(c *gin.Context) {
		role := c.Param("role")
		session := sessions.Default(c)
		session.Set("role", role)
		session.Save()
		c.Status(http.StatusOK)
	})

	r.Use(RoleRequired("admin"))
	r.GET("/admin", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Case 1: Wrong Role
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/login/user", nil)
	r.ServeHTTP(w, req)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/admin", nil)
	req2.Header.Set("Cookie", w.Header().Get("Set-Cookie"))
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusForbidden, w2.Code)

	// Case 2: Correct Role
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/login/admin", nil)
	r.ServeHTTP(w3, req3)

	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("GET", "/admin", nil)
	req4.Header.Set("Cookie", w3.Header().Get("Set-Cookie"))
	r.ServeHTTP(w4, req4)
	assert.Equal(t, http.StatusOK, w4.Code)
}

func TestRateLimiter(t *testing.T) {
	r := setupTestRouter()
	// Strict limit: 1 request per second, burst 1
	r.Use(RateLimiter(rate.Limit(1), 1))
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 1st request: OK
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 2nd request (immediate): Should Fail (Too Many Requests)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRequestLogger(t *testing.T) {
	// Not easy to test log output without redirecting stdout/stderr,
	// but we can ensure it doesn't panic and calls Next()
	r := setupTestRouter()
	r.Use(RequestLogger())
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test Health Check Throttling logic path
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNotFound, w2.Code) // 404 because route not registered, but middleware ran
}
