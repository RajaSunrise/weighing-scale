package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"stoneweigh/internal/models"
	"stoneweigh/internal/pkg/captcha"
)

func TestLogin_Success(t *testing.T) {
	r, db := setupServer(t)

	// Seed User
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := models.User{Username: "admin", PasswordHash: string(hash), Role: "admin"}
	db.Create(&user)

	// Mock Captcha
	captchaID := "test-captcha-valid"
	captchaAnswer := "123456"
	captcha.Store.Set(captchaID, captchaAnswer)

	// Request
	payload := map[string]string{
		"username":   "admin",
		"password":   "password123",
		"captcha_id": captchaID,
		"captcha":    captchaAnswer,
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Login successful")

	// Verify Session Cookie
	assert.NotEmpty(t, w.Header().Get("Set-Cookie"))
}

func TestLogin_InvalidPassword(t *testing.T) {
	r, db := setupServer(t)

	// Seed User
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := models.User{Username: "admin", PasswordHash: string(hash), Role: "admin"}
	db.Create(&user)

	// Mock Captcha
	captchaID := "test-captcha-pass"
	captchaAnswer := "123456"
	captcha.Store.Set(captchaID, captchaAnswer)

	// Request
	payload := map[string]string{
		"username":   "admin",
		"password":   "wrongpassword",
		"captcha_id": captchaID,
		"captcha":    captchaAnswer,
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_InvalidCaptcha(t *testing.T) {
	r, db := setupServer(t)

	// Seed User
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := models.User{Username: "admin", PasswordHash: string(hash), Role: "admin"}
	db.Create(&user)

	// Request with wrong captcha
	payload := map[string]string{
		"username":   "admin",
		"password":   "password123",
		"captcha_id": "any-id",
		"captcha":    "wrong-answer",
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Kode Captcha salah")
}

func TestLogout(t *testing.T) {
	r, _ := setupServer(t)

	// Create a session first (simulated)
	r.GET("/mock_login", func(c *gin.Context) {
		s := sessions.Default(c)
		s.Set("user_id", uint(1))
		s.Save()
		c.Status(200)
	})

	// 1. Login
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/mock_login", nil)
	r.ServeHTTP(w, req)
	cookie := w.Header().Get("Set-Cookie")

	// 2. Logout
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/logout", nil)
	req2.Header.Set("Cookie", cookie)
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusFound, w2.Code)
	// Location should be /login
	assert.Equal(t, "/login", w2.Header().Get("Location"))
}

func TestGetCaptcha(t *testing.T) {
	r, _ := setupServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/captcha", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp["id"])
	assert.NotEmpty(t, resp["b64"])
}
