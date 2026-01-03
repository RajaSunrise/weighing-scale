package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"stoneweigh/internal/handlers"
	"stoneweigh/internal/pkg/templates"
)

// SetupTestRouter initializes a router with the handlers and minimal templates
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Create a minimal dummy server instance
	// We pass nil for dependencies as the public handlers don't need them
	server := &handlers.Server{}

	// Register Public Handlers
	r.GET("/", server.ShowHome)
	r.GET("/produk", server.ShowProduct)
	r.GET("/galeri", server.ShowGallery)

	// Mock Template Loading:
	// Since we can't easily rely on filesystem in unit tests without extensive setup,
	// we will manually load a string template for testing.
	t, _ := templates.LoadTemplatesFromString(`
{{ define "home.html" }}Home{{ end }}
{{ define "produk.html" }}Products{{ end }}
{{ define "galeri.html" }}Gallery{{ end }}
`)
	r.SetHTMLTemplate(t)

	return r
}

func TestPublicRoutes(t *testing.T) {
	r := SetupTestRouter()

	tests := []struct {
		path         string
		expectedCode int
		expectedBody string
	}{
		{"/", 200, "Home"},
		{"/produk", 200, "Products"},
		{"/galeri", 200, "Gallery"},
	}

	for _, tc := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", tc.path, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, tc.expectedCode, w.Code)
		assert.Contains(t, w.Body.String(), tc.expectedBody)
	}
}
