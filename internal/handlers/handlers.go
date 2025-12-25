package handlers

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"stoneweigh/internal/cv"
	"stoneweigh/internal/hardware"
	"stoneweigh/internal/models"
	"stoneweigh/internal/reporting"
)

type Server struct {
	DB          *gorm.DB
	ScaleMgr    *hardware.ScaleManager
	ANPRService *cv.ANPRService
}

func NewServer(db *gorm.DB, sm *hardware.ScaleManager, anpr *cv.ANPRService) *Server {
	return &Server{DB: db, ScaleMgr: sm, ANPRService: anpr}
}

// === VIEW HANDLERS ===

func (s *Server) ShowDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":   "Dashboard",
		"active":  "dashboard",
		"showNav": true,
	})
}

func (s *Server) ShowWeighing(c *gin.Context) {
	c.HTML(http.StatusOK, "weighing.html", gin.H{
		"title":   "Weighing Station",
		"active":  "weighing",
		"showNav": true,
	})
}

// === API HANDLERS ===

// SaveTransaction handles the final weighing and invoice generation
func (s *Server) SaveTransaction(c *gin.Context) {
	var input struct {
		ScaleID     uint    `json:"scale_id"`
		PlateNumber string  `json:"plate_number"`
		DriverName  string  `json:"driver_name"`
		Product     string  `json:"product"`
		Gross       float64 `json:"gross"`
		Tare        float64 `json:"tare"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	net := input.Gross - input.Tare
	ticket := fmt.Sprintf("T-%d", time.Now().Unix())

	record := models.WeighingRecord{
		TicketNumber: ticket,
		ScaleID:      input.ScaleID,
		PlateNumber:  input.PlateNumber,
		DriverName:   input.DriverName,
		Product:      input.Product,
		GrossWeight:  input.Gross,
		TareWeight:   input.Tare,
		NetWeight:    net,
		Status:       "COMPLETED",
		WeighedAt:    time.Now(),
	}

	// Generate PDF
	path, err := reporting.GenerateInvoice(record)
	if err == nil {
		record.InvoicePath = path
	}

	if err := s.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction saved",
		"ticket":  ticket,
		"invoice": record.InvoicePath,
	})
}

// TriggerANPR captures a frame and detects license plate
func (s *Server) TriggerANPR(c *gin.Context) {
	// In a real scenario, we'd map ScaleID to a specific camera URL
	// cameraURL := "rtsp://..."
	cameraURL := "0" // Default webcam

	plate, snapshotPath, err := s.ANPRService.CaptureAndDetect(cameraURL)
	if err != nil {
		// Fallback for demo/simulation if no camera
		c.JSON(http.StatusOK, gin.H{
			"plate": "B 1234 DEMO",
			"snapshot": "/static/images/placeholder_truck.jpg",
			"status": "simulated",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plate":    plate,
		"snapshot": snapshotPath,
		"status":   "success",
	})
}

// StreamScaleData sets up an SSE stream for real-time weights
func (s *Server) StreamScaleData(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// clientChan := make(chan hardware.ScaleData, 10)

	// Determine how to register this client to the broadcater.
	// For simplicity in this MVP, we'll just listen to the main channel
	// (Note: In production, we need a proper fan-out pattern or this will steal messages)
	// We will simulate a dedicated listener or just poll the state for now to avoid complexity of a full Hub.

	// Better approach for MVP: Just loop and fetch current state or wait for a specific event
	// But `hardware.Manager.DataChannel` is a single channel.
	// We need a proper broadcaster.

	// Quick fix: Just send a "connected" message and rely on client polling or
	// actually implement the Hub.
	// Let's implement a simple poll loop for the specific scale requested or all.

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	c.Stream(func(w io.Writer) bool {
		if _, ok := <-ticker.C; ok {
			// Iterate all scales and send data
			// In a real app, we'd only send *changes*
			s.ScaleMgr.Mu.Lock()
			for id, scale := range s.ScaleMgr.Scales {
				c.SSEvent("message", gin.H{
					"scale_id": id,
					"weight":   scale.LastWeight,
					"connected": scale.Connected,
				})
			}
			s.ScaleMgr.Mu.Unlock()
			return true
		}
		return false
	})
}
