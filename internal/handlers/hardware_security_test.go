package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"stoneweigh/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateStation_XSS(t *testing.T) {
	r, server := setupTestServer(t)

	// Register Admin Route
	// Note: In real app this is protected by role check, but here we invoke handler directly or via router with middleware
	// Since setupTestServer sets up sessions, we can use a helper to set admin session

	r.POST("/api/stations", server.CreateStation)

	// Login Helper
	r.POST("/login_admin", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("user_id", uint(1))
		session.Set("role", "admin")
		session.Save()
		c.Status(200)
	})

	wLogin := httptest.NewRecorder()
	reqLogin, _ := http.NewRequest("POST", "/login_admin", nil)
	r.ServeHTTP(wLogin, reqLogin)
	cookie := wLogin.Header().Get("Set-Cookie")

	// XSS Payload in ScalePort
	payload := map[string]any{
		"name":       "Hacker Station",
		"scale_port": "<script>alert(1)</script>",
		"baud_rate":  9600,
		"enabled":    true,
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/stations", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)

	r.ServeHTTP(w, req)

	// EXPECTATION: Should FAIL with 400 Bad Request due to input validation
	// Currently it likely passes (200), so this test will fail until we fix it.
	assert.Equal(t, http.StatusBadRequest, w.Code, "Should reject XSS payload in scale_port")

	// Verify it wasn't created
	var count int64
	server.DB.Model(&models.WeighingStation{}).Where("name = ?", "Hacker Station").Count(&count)
	assert.Equal(t, int64(0), count, "Should not create station with invalid port")
}
