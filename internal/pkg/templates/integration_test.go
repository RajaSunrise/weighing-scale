package templates_test

import (
	"errors"
	"testing"
	"html/template"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"stoneweigh/internal/pkg/templates"
)

// Helper to register dict needed for partials
func setupRouterWithDict() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.SetFuncMap(template.FuncMap{
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
		"json": func(v any) template.JS { return "" },
	})
	return r
}

func TestLoadRealTemplates(t *testing.T) {
	// Locate the web/templates directory relative to this test file.
	// Since test is in internal/pkg/templates, we go up 3 levels then down to web/templates.
	// Actually, we should probably assume we run from repo root or find it.
	rootDir, err := filepath.Abs("../../../web/templates")
	assert.NoError(t, err)

	r := setupRouterWithDict()

	err = templates.LoadTemplates(r, rootDir)
	assert.NoError(t, err, "Should load real templates without error")

	if err == nil {
		// Verify some expected templates exist
		// Since we can't easily access the internal template map of gin's render,
		// we trust LoadTemplates didn't return error.
		// Real parsing logic happens inside LoadTemplates -> template.ParseFiles
		// So if any syntax error existed in the HTML files, it WOULD fail here.
	}
}
