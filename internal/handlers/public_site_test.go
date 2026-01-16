package handlers

import (
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPublicRoutes(t *testing.T) {
	// Setup Gin
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Setup Templates
	// We use a simplified template loader for testing that mimics the app's structure
	tmpl := template.New("")
	tmpl.Funcs(template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"currentYear": func() int {
			return 2026
		},
		"json": func(v any) template.JS {
			return template.JS("{}")
		},
	})

	// Load templates. We need partials and the specific pages.
	// Note: We use relative paths assuming the test is run from internal/handlers
	// This is standard Go testing behavior.
	_, err := tmpl.ParseGlob("../../web/templates/public/partials/*.html")
	if err != nil {
		t.Fatalf("Failed to parse partials: %v", err)
	}
	_, err = tmpl.ParseGlob("../../web/templates/public/*.html")
	if err != nil {
		t.Fatalf("Failed to parse pages: %v", err)
	}
	r.SetHTMLTemplate(tmpl)

	// Mock Server
	server := &Server{}

	// Register Routes to Test
	r.GET("/", server.ShowHome)
	r.GET("/produk", server.ShowProduct)
	r.GET("/galeri", server.ShowGallery)

	// Test Cases
	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedText   string
	}{
		{
			name:           "Home Page",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedText:   "Mitra Batu Split. Hak Cipta Dilindungi Undang-Undang.",
		},
		{
			name:           "Product Page",
			path:           "/produk",
			expectedStatus: http.StatusOK,
			expectedText:   "Spesifikasi Teknis", // Content from produk.html
		},
		{
			name:           "Gallery Page",
			path:           "/galeri",
			expectedStatus: http.StatusOK,
			expectedText:   "Galeri Proyek", // Content from galeri.html
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.path, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedText != "" {
				assert.Contains(t, w.Body.String(), tc.expectedText)
			}
		})
	}
}
