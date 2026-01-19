package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"stoneweigh/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
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

	gin.SetMode(gin.TestMode)
	server := handlers.NewServer(nil, nil, nil)

	// Recover from panics during setup (e.g. bad template path)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Router setup panicked: %v", r)
		}
	}()

	r := SetupRouter(server)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify Security Headers
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "SAMEORIGIN", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}
