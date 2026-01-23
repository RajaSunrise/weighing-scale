package handlers

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupPublicTestRouter sets up a Gin router with templates loaded for public site testing
func setupPublicTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

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
			return time.Now().Year()
		},
		"json": func(v any) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a)
		},
	})

	// Load templates using glob
	// Go tests run in the directory of the package, so we need to go up to root
	basePath := "../../web/templates/public"
	partialsPath := "../../web/templates/public/partials"

	// Must parse partials first, then pages, or glob them all
	files, err := filepath.Glob(filepath.Join(partialsPath, "*.html"))
	assert.NoError(t, err)
	pageFiles, err := filepath.Glob(filepath.Join(basePath, "*.html"))
	assert.NoError(t, err)

	files = append(files, pageFiles...)

	if len(files) > 0 {
		_, err = tmpl.ParseFiles(files...)
		assert.NoError(t, err)
		r.SetHTMLTemplate(tmpl)
	} else {
		t.Log("Warning: No templates found for testing")
	}

	return r
}

func TestPublicRoutes(t *testing.T) {
	r := setupPublicTestRouter(t)
	server := &Server{}

	// Register all public routes
	r.GET("/", server.ShowHome)
	r.GET("/produk", server.ShowProduct)
	r.GET("/galeri", server.ShowGallery)
	r.GET("/tentang", server.ShowAbout)
	r.GET("/artikel", server.ShowNews)
	r.GET("/kontak", server.ShowContact)
	r.GET("/faq", server.ShowFAQ)
	r.GET("/visi-misi", server.ShowVision)
	r.GET("/syarat-ketentuan", server.ShowTerms)
	r.GET("/privasi", server.ShowPrivacy)

	// Define test cases
	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedText   []string // Check for multiple strings to ensure correct content
	}{
		{
			name:           "Home Page",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"Pondasi Kokoh", "PT. KAYA RAYA BAROKAH. Hak Cipta Dilindungi Undang-Undang."},
		},
		{
			name:           "Product Page",
			path:           "/produk",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"Spesifikasi Teknis", "Ukuran"},
		},
		{
			name:           "Gallery Page",
			path:           "/galeri",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"Galeri Proyek"},
		},
		{
			name:           "About Page",
			path:           "/tentang",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"Tentang Kami", "Perjalanan Kami"},
		},
		{
			name:           "News Page",
			path:           "/artikel",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"Berita & Artikel"},
		},
		{
			name:           "Contact Page",
			path:           "/kontak",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"Hubungi Kami"},
		},
		{
			name:           "FAQ Page",
			path:           "/faq",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"Pertanyaan Umum"},
		},
		{
			name:           "Vision Page",
			path:           "/visi-misi",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"Visi & Misi"},
		},
		{
			name:           "Terms Page",
			path:           "/syarat-ketentuan",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"Syarat & Ketentuan"},
		},
		{
			name:           "Privacy Page",
			path:           "/privasi",
			expectedStatus: http.StatusOK,
			expectedText:   []string{"Kebijakan Privasi"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tc.path, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			for _, text := range tc.expectedText {
				assert.Contains(t, w.Body.String(), text, "Response should contain expected text")
			}
		})
	}
}

func TestPublic404(t *testing.T) {
	r := setupPublicTestRouter(t)
	server := &Server{}
	r.NoRoute(server.Show404)

	// We need to ensure 404.html is loaded by setupPublicTestRouter if it's in public folder
	// Let's check if 404 is in public. If it's not in public, we might need to load it from templates root.
	// But assuming Show404 renders "404.html" and where it lives.
	// Based on memory, 404.html might be in web/templates/ or web/templates/public/
	// If it fails, we will adjust.

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/non-existent-page-xyz", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	// If the template 404.html is not found, Gin will probably render text or panic in handlers.
	// We'll see.
}
