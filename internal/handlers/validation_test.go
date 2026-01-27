package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"stoneweigh/internal/models"
)

func TestSaveTransaction_InputLength(t *testing.T) {
	r, server := setupTestServer(t)
	// Register route
	r.POST("/api/transaction", server.SaveTransaction)

	db := server.DB

	// Setup Data
	station := models.WeighingStation{Name: "Station A", Enabled: true}
	db.Create(&station)

	// User
	user := models.User{Username: "operator", Role: "operator"}
	db.Create(&user)

	// Assignment
	assignment := models.UserStationAssignment{UserID: user.ID, WeighingStationID: station.ID}
	db.Create(&assignment)

	// Helper to login
	r.POST("/login_test_len", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("user_id", user.ID)
		session.Set("username", user.Username)
		session.Set("role", user.Role)
		session.Save()
		c.Status(200)
	})

	// Login
	wLogin := httptest.NewRecorder()
	reqLogin, _ := http.NewRequest("POST", "/login_test_len", nil)
	r.ServeHTTP(wLogin, reqLogin)
	cookie := wLogin.Header().Get("Set-Cookie")

	// Create massive string (10KB)
	massiveString := strings.Repeat("A", 10240)

	// Test Case: Massive Plate Number
	payload := map[string]any{
		"scale_id":     station.ID,
		"plate_number": massiveString,
		"driver_name":  "Driver OK",
		"gross":        10000,
		"tare":         5000,
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/transaction", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	r.ServeHTTP(w, req)

	// Should fail with 400
	assert.Equal(t, http.StatusBadRequest, w.Code, "Should reject massive input")

	// Verify error message if possible
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if msg, ok := resp["error"].(string); ok {
		assert.Contains(t, msg, "PlateNumber", "Error should mention the invalid field")
	}
}
