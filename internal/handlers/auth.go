package handlers

import (
	"html/template"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
	"golang.org/x/crypto/bcrypt"
	"stoneweigh/internal/models"
	"stoneweigh/internal/pkg/captcha"
)

// Pre-calculated bcrypt hash (cost 10) for "dummy" to prevent timing attacks
const dummyHash = "$2a$10$dvlxIvfO4J9B9HPYx32SoO9/pfX.sQ7r/4Eu/16EHhhUh3S4PlYmq"

// ShowLogin renders the login page
func (s *Server) ShowLogin(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("user_id") != nil {
		c.Redirect(http.StatusFound, "/dashboard")
		return
	}

	// Generate Captcha
	captchaID, captchaB64, err := captcha.GenerateCaptcha()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"error": "Failed to generate captcha"})
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"title":      "Login",
		"csrf_token": csrf.GetToken(c),
		"captchaID":  captchaID,
		"captchaB64": template.URL(captchaB64),
	})
}

// GetCaptcha returns a new captcha via JSON
func (s *Server) GetCaptcha(c *gin.Context) {
	id, b64, err := captcha.GenerateCaptcha()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate captcha"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "b64": b64})
}

// Login handles the authentication
func (s *Server) Login(c *gin.Context) {
	var input struct {
		Username  string `json:"username" binding:"required"`
		Password  string `json:"password" binding:"required"`
		CaptchaID string `json:"captcha_id" binding:"required"`
		Captcha   string `json:"captcha" binding:"required"`
	}

	// Support both JSON and Form (for simple HTML login)
	if c.ContentType() == "application/json" {
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
	} else {
		input.Username = c.PostForm("username")
		input.Password = c.PostForm("password")
		input.CaptchaID = c.PostForm("captcha_id")
		input.Captcha = c.PostForm("captcha")
	}

	// Verify Captcha
	if !captcha.VerifyCaptcha(input.CaptchaID, input.Captcha) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kode Captcha salah"})
		return
	}

	var user models.User
	err := s.DB.Where("username = ?", input.Username).First(&user).Error

	// Timing attack mitigation:
	// Always perform password comparison, even if user is not found.
	// This prevents attackers from enumerating valid usernames by measuring response time.
	hashToCompare := user.PasswordHash
	if err != nil {
		hashToCompare = dummyHash
	}

	errPassword := bcrypt.CompareHashAndPassword([]byte(hashToCompare), []byte(input.Password))

	if err != nil || errPassword != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Create Session
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("username", user.Username)
	session.Set("full_name", user.FullName)
	session.Set("role", user.Role)
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "redirect": "/dashboard"})
}

// Logout destroys the session
func (s *Server) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/login")
}
