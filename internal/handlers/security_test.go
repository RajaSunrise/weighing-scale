package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

	dsn := fmt.Sprintf("file:memdb_sec%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.User{},
		&models.WeighingStation{},
		&models.StationCamera{},
		&models.UserStationAssignment{},
		&models.WeighingRecord{},
	)
	assert.NoError(t, err)

	// Mock dependencies
	hardware.InitScaleManager()
	server := NewServer(db, hardware.Manager, cv.NewANPRService("mock.onnx"), nil)

	r := gin.New()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// Mock CSRF Token (needed for some handlers)
	r.Use(func(c *gin.Context) {
		c.Set("csrfSecret", "test-secret")
		c.Next()
	})

	return r, server
}

func TestSaveTransaction_IDOR(t *testing.T) {
	r, server := setupTestServer(t)
	// Register route
	r.POST("/api/transaction", server.SaveTransaction)

	db := server.DB

	// Setup Data
	// 1. Create Stations
	stationAllowed := models.WeighingStation{Name: "Station A", Enabled: true}
	stationForbidden := models.WeighingStation{Name: "Station B", Enabled: true}
	db.Create(&stationAllowed)
	db.Create(&stationForbidden)

	// 2. Create User
	user := models.User{Username: "operator", Role: "operator"}
	db.Create(&user)

	// 3. Assign User to Station A ONLY
	assignment := models.UserStationAssignment{UserID: user.ID, WeighingStationID: stationAllowed.ID}
	db.Create(&assignment)

	// Helper to login as operator
	r.POST("/login_test_idor", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("user_id", user.ID)
		session.Set("username", user.Username)
		session.Set("role", user.Role)
		session.Save()
		c.Status(200)
	})

	// Establish Session
	wLogin := httptest.NewRecorder()
	reqLogin, _ := http.NewRequest("POST", "/login_test_idor", nil)
	r.ServeHTTP(wLogin, reqLogin)
	cookie := wLogin.Header().Get("Set-Cookie")

	// Test Case 1: Access Allowed Station (Should Success)
	payloadAllowed := map[string]any{
		"scale_id":     stationAllowed.ID,
		"plate_number": "B 1234 OK",
		"driver_name":  "Driver OK",
		"gross":        10000,
		"tare":         5000,
	}
	bodyAllowed, _ := json.Marshal(payloadAllowed)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/transaction", strings.NewReader(string(bodyAllowed)))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Cookie", cookie)
	r.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code, "Operator should be able to write to assigned station")

	// Test Case 2: Access Forbidden Station (Should Fail with 403)
	payloadForbidden := map[string]any{
		"scale_id":     stationForbidden.ID,
		"plate_number": "B 666 BAD",
		"driver_name":  "Driver BAD",
		"gross":        10000,
		"tare":         5000,
	}
	bodyForbidden, _ := json.Marshal(payloadForbidden)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/transaction", strings.NewReader(string(bodyForbidden)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Cookie", cookie)
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusForbidden, w2.Code, "Operator should NOT be able to write to unassigned station")
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
	var resp1 map[string]any
	json.Unmarshal(w1.Body.Bytes(), &resp1)
	assert.Contains(t, resp1["error"], "at least 8 characters")

	// Case 2: Weak Password (No Number)
	payloadWeak2 := `{"username": "nonumber", "password": "password", "full_name": "No Number", "role": "operator"}`
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/users", strings.NewReader(payloadWeak2))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)
	var resp2 map[string]any
	json.Unmarshal(w2.Body.Bytes(), &resp2)
	assert.Contains(t, resp2["error"], "at least one number")

	// Case 3: Strong Password
	payloadStrong := `{"username": "strong", "password": "Password123", "full_name": "Strong", "role": "operator"}`
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/api/users", strings.NewReader(payloadStrong))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)

	assert.Equal(t, http.StatusOK, w3.Code)
	var resp3 map[string]any
	json.Unmarshal(w3.Body.Bytes(), &resp3)
	assert.Equal(t, "User created", resp3["message"])

	// Case 4: Weak Password (No Uppercase)
	payloadWeak3 := `{"username": "noupper", "password": "password123", "full_name": "No Upper", "role": "operator"}`
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest("POST", "/api/users", strings.NewReader(payloadWeak3))
	req4.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w4, req4)

	assert.Equal(t, http.StatusBadRequest, w4.Code)
	var resp4 map[string]any
	json.Unmarshal(w4.Body.Bytes(), &resp4)
	assert.Contains(t, resp4["error"], "uppercase letter")
}
