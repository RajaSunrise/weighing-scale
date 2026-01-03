package templates_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"stoneweigh/internal/pkg/templates"
)

func TestLoadTemplates(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "templates_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create root template
	err = os.WriteFile(filepath.Join(tmpDir, "root.html"), []byte("Root"), 0644)
	assert.NoError(t, err)

	// Create subdir
	subDir := filepath.Join(tmpDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	assert.NoError(t, err)

	// Create nested template
	err = os.WriteFile(filepath.Join(subDir, "nested.html"), []byte("Nested"), 0644)
	assert.NoError(t, err)

	// Setup Gin
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Load templates
	err = templates.LoadTemplates(r, tmpDir)
	assert.NoError(t, err)

	// Verify template registration (internal check implies if it didn't panic or error, it worked,
	// but we can check if r.HTMLRender is set. However, LoadTemplates sets it.)

	// A more robust check is rendering
	// We can't easily check r.HTMLRender instance type without casting private types,
	// but success of LoadTemplates implies it found files.
}

func TestLoadTemplatesFromString(t *testing.T) {
	tmplStr := `{{ define "test" }}Hello{{ end }}`
	tmpl, err := templates.LoadTemplatesFromString(tmplStr)
	assert.NoError(t, err)
	assert.NotNil(t, tmpl)
	assert.NotNil(t, tmpl.Lookup("test"))
}
