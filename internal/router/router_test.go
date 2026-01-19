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

func TestPublicRoute(t *testing.T) {
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

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	// Check for specific content from home.html
	assert.Contains(t, w.Body.String(), "Pondasi Kokoh")

	// Check for dynamic footer year
	assert.Contains(t, w.Body.String(), "Mitra Batu Split. Hak Cipta Dilindungi Undang-Undang.")
}
