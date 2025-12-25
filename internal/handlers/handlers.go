package handlers

import (
	"net/http"
	"time"
	"math/rand"
	"strconv"

	"github.com/gin-gonic/gin"
	"stoneweigh/internal/models"
	"stoneweigh/internal/database"
)

func ShowLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":     "Login",
		"component": "LoginScreen",
		"showNav":   false,
	})
}

func ShowDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":     "Dashboard",
		"component": "DashboardScreen",
		"showNav":   true,
	})
}

func ShowWeighing(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"title":     "Weighing Station",
		"component": "WeighingStationScreen",
		"showNav":   true,
	})
}

// API Handlers (mock implementation for some parts)

func Login(c *gin.Context) {
	// Mock login
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Real implementation would check DB
	// For prototype, accept any non-empty
	if creds.Email != "" && creds.Password != "" {
		c.JSON(http.StatusOK, gin.H{"token": "mock-token-123"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}
}

func GetTransactions(c *gin.Context) {
	var transactions []models.Transaction
	// Check if DB is connected
	if database.DB != nil {
		database.DB.Find(&transactions)
	}
	c.JSON(http.StatusOK, transactions)
}

func CreateTransaction(c *gin.Context) {
	var txn models.Transaction
	if err := c.ShouldBindJSON(&txn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	txn.TicketID = "T-" + strconv.Itoa(rand.Intn(100000))
	txn.EntryTime = time.Now()
	txn.Status = "PENDING"

	if database.DB != nil {
		database.DB.Create(&txn)
	}
	c.JSON(http.StatusOK, txn)
}

// Mock CCTV Stream Handler (MJPEG)
func StreamCCTV(c *gin.Context) {
	// In a real app, this would stream MJPEG frames from gocv
	c.String(http.StatusOK, "CCTV Stream Placeholder")
}
