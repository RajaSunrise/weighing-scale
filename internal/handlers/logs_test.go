package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetLogsAPI(t *testing.T) {
	// Setup environment
	logDir := "logs"
	logFile := filepath.Join(logDir, "system.log")

	// Ensure cleanup first in case previous run failed
	os.RemoveAll(logDir)

	err := os.MkdirAll(logDir, 0755)
	assert.NoError(t, err)
	defer os.RemoveAll(logDir)

	server := &Server{}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/logs", server.GetLogsAPI)

	t.Run("Standard Large File", func(t *testing.T) {
		// Create dummy log file with 150 lines
		f, err := os.Create(logFile)
		assert.NoError(t, err)

		var expectedLast100 []string
		for i := 1; i <= 150; i++ {
			line := fmt.Sprintf("Log line %d", i)
			f.WriteString(line + "\n")
			if i > 50 {
				expectedLast100 = append(expectedLast100, line)
			}
		}
		f.Close()

		// Perform Request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/logs", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []string
		err = json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)

		assert.Equal(t, 100, len(result))
		assert.Equal(t, expectedLast100, result)
	})

	tests := []struct {
		name          string
		content       string
		expectedCount int
		expectedLast  string
		expectedFirst string // Check first returned line
	}{
		{
			name:          "Empty File",
			content:       "",
			expectedCount: 0,
		},
		{
			name:          "Small File",
			content:       "Line1\nLine2\n",
			expectedCount: 2,
			expectedLast:  "Line2",
			expectedFirst: "Line1",
		},
		{
			name:          "Exactly 100 Lines",
			content:       generateLines(100),
			expectedCount: 100,
			expectedLast:  "Line 100",
			expectedFirst: "Line 1",
		},
		{
			name:          "More than 100 Lines",
			content:       generateLines(150),
			expectedCount: 100,
			expectedLast:  "Line 150",
			expectedFirst: "Line 51",
		},
		{
			name:          "Single Newline",
			content:       "\n",
			expectedCount: 1, // Should return [""]
			expectedLast:  "",
		},
		{
			name:          "Trailing Newline Handling",
			content:       "A\nB\n",
			expectedCount: 2,
			expectedLast:  "B",
			expectedFirst: "A",
		},
		{
			name:          "No Trailing Newline",
			content:       "A\nB",
			expectedCount: 2,
			expectedLast:  "B",
			expectedFirst: "A",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Write content
			err := os.WriteFile(logFile, []byte(tc.content), 0644)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/logs", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var result []string
			err = json.Unmarshal(w.Body.Bytes(), &result)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedCount, len(result), "Count mismatch for %s", tc.name)

			if tc.expectedCount > 0 {
				if tc.expectedLast != "" {
					assert.Equal(t, tc.expectedLast, result[len(result)-1])
				}
				if tc.expectedFirst != "" {
					assert.Equal(t, tc.expectedFirst, result[0])
				}
			}

			if tc.name == "Single Newline" {
				assert.Equal(t, "", result[0])
			}
		})
	}
}

func generateLines(n int) string {
	var b bytes.Buffer
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "Line %d\n", i)
	}
	return b.String()
}
