package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"stoneweigh/internal/cv"
	"stoneweigh/internal/hardware"
	"stoneweigh/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestServer exposes the server instance
func setupTestServer(t *testing.T) (*gin.Engine, *Server) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.User{})
	assert.NoError(t, err)

	// Mock dependencies
	hardware.InitScaleManager()
	server := NewServer(db, hardware.Manager, cv.NewANPRService("mock.onnx"))

	r := gin.New()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	return r, server
}

func TestCreateUser_PasswordPolicy(t *testing.T) {
	r, server := setupTestServer(t)
	r.POST("/api/users", server.CreateUser)

	// Case 1: Weak Password (Too Short)
	payloadWeak1 := `{"username": "short", "password": "123", "full_name": "Short", "role": "operator"}`
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/users", strings.NewReader(payloadWeak1))
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusBadRequest, w1.Code)
	var resp1 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &resp1)
	assert.Contains(t, resp1["error"], "at least 8 characters")

	// Case 2: Weak Password (No Number)
	payloadWeak2 := `{"username": "nonumber", "password": "password", "full_name": "No Number", "role": "operator"}`
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/users", strings.NewReader(payloadWeak2))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)
	var resp2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp2)
	assert.Contains(t, resp2["error"], "at least one number")

	// Case 3: Strong Password
	payloadStrong := `{"username": "strong", "password": "Password123", "full_name": "Strong", "role": "operator"}`
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/api/users", strings.NewReader(payloadStrong))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)
	var resp3 map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &resp3)
	assert.Equal(t, "User created", resp3["message"])
}
