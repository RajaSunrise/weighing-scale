package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"strings"

	"stoneweigh/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLoginRateLimit(t *testing.T) {
	// Robust way to find templates: walk up until we find web/templates
	wd, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(wd, "web", "templates")); err == nil {
			os.Chdir(wd)
			break
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}

	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Create a server with nil dependencies
	// This will cause handlers to error/panic, but middleware runs before handlers.
	server := handlers.NewServer(nil, nil, nil)

	r := SetupRouter(server)

	// We will simulate 10 requests.
	// The proposed rate limit is 1 req/min with burst of 5.
	// So requests 1-5 should pass (return 400 or 500 because nil DB/Captcha).
	// Requests 6+ should be blocked (return 429).

	for i := 1; i <= 10; i++ {
		w := httptest.NewRecorder()
		// Using POST /login
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(`{"username":"admin","password":"password"}`))
		req.Header.Set("Content-Type", "application/json")

		// Important: We need to mock a unique IP, OR reuse the same IP to trigger rate limiting.
		// httptest requests don't have a remote address by default.
		// Gin's ClientIP() uses RemoteAddr.
		req.RemoteAddr = "192.168.1.100:1234"

		r.ServeHTTP(w, req)

		if i <= 5 {
			// First 5 requests should be allowed (and fail due to invalid captcha/nil DB)
			// We expect 400 (Bad Request - Invalid Input/Captcha) or 500 (Internal Error)
			// But DEFINITELY NOT 429.
			if w.Code == http.StatusTooManyRequests {
				t.Fatalf("Request %d was rate limited prematurely", i)
			}
		} else {
			// Requests 6+ should be rate limited
			assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request %d should be rate limited. Body: %s", i, w.Body.String())
		}
	}
}
