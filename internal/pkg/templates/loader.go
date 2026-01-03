package templates

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// LoadTemplates loads all HTML templates from the specified root directory and its subdirectories.
// It bypasses the limitation of filepath.Glob which doesn't support recursive "**" matching in standard Go.
func LoadTemplates(r *gin.Engine, rootDir string) error {
	var files []string

	// Walk the directory tree to find all .html files
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk templates directory: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no templates found in %s", rootDir)
	}

	// Create a new template and parse all found files
	// We use the base name of the file as the template name,
	// UNLESS the file itself defines a template name (which Go templates handle automatically).
	// For files in subdirectories that don't use {{ define }}, they are usually accessed by filename.
	// However, collisions can occur (e.g., templates/index.html vs templates/public/index.html).
	// Go's ParseFiles uses the base name.
	// To support distinct names for subdirectories, we might need a custom loader.
	// BUT: The existing app uses {{ define "header" }} etc.
	// The new public pages are standalone files.
	// We can parse them all. If there are duplicate filenames (e.g. index.html in root and public),
	// the last one parsed wins if we use ParseFiles on all of them into one set.

	// Strategy:
	// Use gin's SetHTMLTemplate.
	// We need to construct a *template.Template that contains all of them.
	// To avoid collisions, we should probably check if we can name them differently.
	// Existing app templates: dashboard.html, etc. (do they use define? yes, header/footer do. dashboard.html uses {{ template "header" }} but does it define itself?
	// Checking dashboard.html: It starts with {{ template "header" . }}. It DOES NOT have a {{ define "dashboard" }}.
	// So it is registered as "dashboard.html".

	// New public templates: index.html, produk.html.
	// If we have "index.html" in root? No, we don't.
	// We have "dashboard.html" in root.
	// So "public/index.html" -> registered as "index.html".
	// This seems safe for now provided no name collisions.

	tmpl := template.New("")
	// We must register the FuncMap BEFORE parsing
	tmpl.Funcs(r.FuncMap)

	// Parse files
	_, err = tmpl.ParseFiles(files...)
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	r.SetHTMLTemplate(tmpl)
	return nil
}

// LoadTemplatesFromString loads templates from a string for testing purposes.
func LoadTemplatesFromString(templateString string) (*template.Template, error) {
	return template.New("").Parse(templateString)
}
