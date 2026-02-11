package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"stoneweigh/internal/handlers"
	"github.com/stretchr/testify/assert"
)

func TestSecurityFixes(t *testing.T) {
	// Setup
	os.Setenv("SESSION_SECRET", "testsecret")
	server := &handlers.Server{} // Mock or empty server for routing tests
	r := SetupRouter(server)

	t.Run("CSRF Bypass via Path Traversal", func(t *testing.T) {
		// Testing bypass attempt: /api/external//../api/transaction
		// Without path.Clean, this might have started with /api/external/
		// But Gin's router will also clean it or fail to match.
		// We want to ensure it's NOT skipped if it resolves outside /api/external.

		req, _ := http.NewRequest("POST", "/api/external/../api/transaction", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		// If skipped, it would go to handlers and probably 401 (if no session)
		// If NOT skipped, it fails CSRF check with 400.
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Contains(t, resp.Body.String(), "CSRF token mismatch")
	})

	t.Run("Cache-Control for Sensitive Reports", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/static/reports/inv_T-123.pdf", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, "private, no-store, must-revalidate", resp.Header().Get("Cache-Control"))
		assert.Equal(t, "no-cache", resp.Header().Get("Pragma"))
	})

	t.Run("Cache-Control for Snapshots", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/static/images/snap_12345.jpg", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, "private, no-store, must-revalidate", resp.Header().Get("Cache-Control"))
	})

	t.Run("Cache-Control for Public Assets", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/static/css/main.css", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, "public, max-age=86400", resp.Header().Get("Cache-Control"))
	})
}
