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

func TestSecurity_InputValidation_XSS(t *testing.T) {
	r, server := setupTestServer(t)

	// Register routes
	r.POST("/api/transaction", server.SaveTransaction)
	r.POST("/api/vehicles", server.CreateVehicle)

	// Admin routes need role middleware usually, but we are testing the handler logic directly here
	// assuming middleware passes or we bypass it for unit testing logic validation.
	// Actually, `CreateVehicle` in router is protected. But here we call handler directly via router without middleware
	// UNLESS setupTestServer sets up middleware.
	// `setupTestServer` usually sets up a basic router.
	// Let's check if we need to mock authentication.

	db := server.DB

	// Setup Data
	station := models.WeighingStation{Name: "Station A", Enabled: true}
	db.Create(&station)

	user := models.User{Username: "operator", Role: "operator"}
	db.Create(&user)

	assignment := models.UserStationAssignment{UserID: user.ID, WeighingStationID: station.ID}
	db.Create(&assignment)

	// Helper to login as operator
	r.POST("/login_test_xss", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("user_id", user.ID)
		session.Set("username", user.Username)
		session.Set("role", user.Role)
		session.Save()
		c.Status(200)
	})

	// Login
	wLogin := httptest.NewRecorder()
	reqLogin, _ := http.NewRequest("POST", "/login_test_xss", nil)
	r.ServeHTTP(wLogin, reqLogin)
	cookie := wLogin.Header().Get("Set-Cookie")

	// 1. Test SaveTransaction with XSS in DriverName
	t.Run("SaveTransaction_XSS", func(t *testing.T) {
		payload := map[string]any{
			"scale_id":     station.ID,
			"plate_number": "B 1234 XY",
			"driver_name":  "<script>alert(1)</script>", // XSS Payload
			"gross":        10000,
			"tare":         5000,
			"company":      "Safe Company",
			"product":      "Sand",
		}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/transaction", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "DriverName")
		assert.Contains(t, w.Body.String(), "invalid characters")
	})

	// 2. Test SaveTransaction with valid input
	t.Run("SaveTransaction_Valid", func(t *testing.T) {
		payload := map[string]any{
			"scale_id":     station.ID,
			"plate_number": "B 1234 XY",
			"driver_name":  "John Doe",
			"gross":        10000,
			"tare":         5000,
			"company":      "Safe Company",
			"product":      "Sand",
		}
		body, _ := json.Marshal(payload)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/transaction", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSecurity_UserValidation(t *testing.T) {
	r, server := setupTestServer(t)
	r.POST("/api/users", server.CreateUser) // Usually admin only

	t.Run("CreateUser_BadUsername", func(t *testing.T) {
		payload := map[string]string{
			"username":  "user name", // Space not allowed
			"password":  "StrongPass1",
			"full_name": "User Name",
			"role":      "operator",
		}
		body, _ := json.Marshal(payload)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/users", strings.NewReader(string(body)))
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateUser_BadRole", func(t *testing.T) {
		payload := map[string]string{
			"username":  "hacker",
			"password":  "StrongPass1",
			"full_name": "Hacker",
			"role":      "superuser", // Invalid role
		}
		body, _ := json.Marshal(payload)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/users", strings.NewReader(string(body)))
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateUser_XSS_FullName", func(t *testing.T) {
		payload := map[string]string{
			"username":  "gooduser",
			"password":  "StrongPass1",
			"full_name": "<img src=x>",
			"role":      "operator",
		}
		body, _ := json.Marshal(payload)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/users", strings.NewReader(string(body)))
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestSecurity_WeightValidation(t *testing.T) {
	r, server := setupTestServer(t)
	// Register route
	r.POST("/api/transaction", server.SaveTransaction)

	db := server.DB

	// Setup Data
	station := models.WeighingStation{Name: "Station Weight", Enabled: true}
	db.Create(&station)

	user := models.User{Username: "operator_w", Role: "operator"}
	db.Create(&user)

	assignment := models.UserStationAssignment{UserID: user.ID, WeighingStationID: station.ID}
	db.Create(&assignment)

	// Helper to login
	r.POST("/login_test_weight", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("user_id", user.ID)
		session.Set("username", user.Username)
		session.Set("role", user.Role)
		session.Save()
		c.Status(200)
	})

	// Login
	wLogin := httptest.NewRecorder()
	reqLogin, _ := http.NewRequest("POST", "/login_test_weight", nil)
	r.ServeHTTP(wLogin, reqLogin)
	cookie := wLogin.Header().Get("Set-Cookie")

	// Helper for making requests
	makeRequest := func(gross, tare float64) *httptest.ResponseRecorder {
		payload := map[string]any{
			"scale_id":     station.ID,
			"plate_number": "B 1234 WGT",
			"driver_name":  "Driver Weight",
			"gross":        gross,
			"tare":         tare,
			"company":      "Weight Co",
			"product":      "Rocks",
		}
		body, _ := json.Marshal(payload)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/transaction", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)
		r.ServeHTTP(w, req)
		return w
	}

	t.Run("Negative Gross", func(t *testing.T) {
		w := makeRequest(-100, 50)
		assert.Equal(t, http.StatusBadRequest, w.Code, "Should reject negative gross")
	})

	t.Run("Negative Tare", func(t *testing.T) {
		w := makeRequest(100, -50)
		assert.Equal(t, http.StatusBadRequest, w.Code, "Should reject negative tare")
	})

	t.Run("Tare Greater Than Gross", func(t *testing.T) {
		w := makeRequest(100, 150)
		assert.Equal(t, http.StatusBadRequest, w.Code, "Should reject tare > gross")
	})

	t.Run("Valid Weights", func(t *testing.T) {
		w := makeRequest(100, 50)
		assert.Equal(t, http.StatusOK, w.Code, "Should accept valid weights")
	})
}
