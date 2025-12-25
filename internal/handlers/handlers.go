package handlers

import (
	"net/http"
	"time"
	"math/rand"
	"strconv"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"stoneweigh/internal/models"
	"stoneweigh/internal/database"
	"stoneweigh/internal/pkg/report"
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

// API Handlers

func Login(c *gin.Context) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Mock auth
	if creds.Email != "" && creds.Password != "" {
		c.JSON(http.StatusOK, gin.H{"token": "mock-token-123"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	}
}

func GetTransactions(c *gin.Context) {
	var transactions []models.Transaction
	if database.DB != nil {
		database.DB.Order("created_at desc").Limit(10).Find(&transactions)
	}
	c.JSON(http.StatusOK, transactions)
}

func GetStats(c *gin.Context) {
	var stats struct {
		TotalCount     int64   `json:"total_count"`
		TotalWeight    float64 `json:"total_weight"`
		PendingCount   int64   `json:"pending_count"`
	}

	if database.DB != nil {
		database.DB.Model(&models.Transaction{}).Count(&stats.TotalCount)
		database.DB.Model(&models.Transaction{}).Select("sum(net_weight)").Scan(&stats.TotalWeight)
		database.DB.Model(&models.Transaction{}).Where("status = ?", "PENDING").Count(&stats.PendingCount)
	}

	c.JSON(http.StatusOK, stats)
}

func CreateTransaction(c *gin.Context) {
	var txn models.Transaction
	if err := c.ShouldBindJSON(&txn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	txn.TicketID = "T-" + strconv.Itoa(rand.Intn(100000))
	txn.EntryTime = time.Now()
	txn.Status = "COMPLETED" // Assume completed for this flow

	if database.DB != nil {
		database.DB.Create(&txn)
	}

	// Generate PDF
	pdfPath, err := report.GenerateTicketPDF(txn)
	pdfUrl := ""
	if err == nil {
		// Convert path to URL
		filename := filepath.Base(pdfPath)
		pdfUrl = "/static/reports/" + filename
	}

	c.JSON(http.StatusOK, gin.H{
		"ticket_id": txn.TicketID,
		"pdf_url":   pdfUrl,
	})
}

// Mock CCTV Stream Handler
func StreamCCTV(c *gin.Context) {
	c.String(http.StatusOK, "CCTV Stream Placeholder")
}
