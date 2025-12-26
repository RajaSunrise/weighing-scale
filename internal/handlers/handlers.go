package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
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
	session := sessions.Default(c)
	fullName := "Operator"
	if v := session.Get("full_name"); v != nil {
		fullName = v.(string)
	}

	// 1. Fetch Stats for Today
	startOfDay := time.Now().Truncate(24 * time.Hour)

	var todayCount int64
	var todayWeight float64 // Sum of NetWeight

	s.DB.Model(&models.WeighingRecord{}).
		Where("weighed_at >= ?", startOfDay).
		Count(&todayCount)

	type Result struct {
		Total float64
	}
	var res Result
	s.DB.Model(&models.WeighingRecord{}).
		Select("sum(net_weight) as total").
		Where("weighed_at >= ?", startOfDay).
		Scan(&res)
	todayWeight = res.Total

	// 2. Fetch Recent Transactions
	var recent []models.WeighingRecord
	s.DB.Order("weighed_at desc").Limit(10).Find(&recent)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":       "Dashboard",
		"active":      "dashboard",
		"showNav":     true,
		"CurrentUser": fullName,
		"Stats": gin.H{
			"TodayCount":  todayCount,
			"TodayWeight": todayWeight,
		},
		"Recent": recent,
	})
}

func (s *Server) ShowWeighing(c *gin.Context) {
	session := sessions.Default(c)
	uidVal := session.Get("user_id")

	fullName := "Operator"
	if v := session.Get("full_name"); v != nil {
		fullName = v.(string)
	}

	// If admin, show all active stations
	// If operator, show only assigned stations
	// Note: We need to pass the list of allowed stations to the template so JS can render them dynamically
	// instead of hardcoded 1,2,3.

	var allowedStations []models.WeighingStation

	if role := session.Get("role"); role == "admin" {
		s.DB.Preload("Cameras").Where("enabled = ?", true).Find(&allowedStations)
	} else if uidVal != nil {
		var assignments []models.UserStationAssignment
		s.DB.Preload("WeighingStation.Cameras").Where("user_id = ?", uidVal).Find(&assignments)
		for _, a := range assignments {
			if a.WeighingStation.Enabled {
				allowedStations = append(allowedStations, a.WeighingStation)
			}
		}
	}

	c.HTML(http.StatusOK, "weighing.html", gin.H{
		"title":       "Weighing Station",
		"active":      "weighing",
		"showNav":     true,
		"CurrentUser": fullName,
		"Stations":    allowedStations,
	})
}

// === API HANDLERS ===

// SaveTransaction handles the final weighing and invoice generation
func (s *Server) SaveTransaction(c *gin.Context) {
	var input struct {
		ScaleID     uint    `json:"scale_id"`
		PlateNumber string  `json:"plate_number"`
		DriverName  string  `json:"driver_name"`
		Company     string  `json:"company"`
		Product     string  `json:"product"`
		Gross       float64 `json:"gross"`
		Tare        float64 `json:"tare"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session := sessions.Default(c)
	managerName := "Unknown"
	if val := session.Get("username"); val != nil {
		managerName = val.(string)
	}

	net := input.Gross - input.Tare
	// Use UnixNano to prevent collision on rapid submissions
	ticket := fmt.Sprintf("T-%d", time.Now().UnixNano())

	record := models.WeighingRecord{
		TicketNumber: ticket,
		ScaleID:      input.ScaleID,
		PlateNumber:  input.PlateNumber,
		DriverName:   input.DriverName,
		CompanyName:  input.Company,
		ManagerName:  managerName,
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
	} else {
		fmt.Printf("Error generating PDF: %v\n", err)
	}

	if err := s.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save record"})
		return
	}

	// Fix PDF Path for Frontend:
	// The reporting package returns relative path like "web/static/reports/..."
	// We need to strip "web" so it becomes "/static/reports/..."
	webPath := "/" + strings.TrimPrefix(record.InvoicePath, "web/")

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction saved",
		"ticket":  ticket,
		"invoice": webPath,
	})
}

// TriggerANPR captures a frame and detects license plate
func (s *Server) TriggerANPR(c *gin.Context) {
	scaleID := c.Query("scale_id")
	camID := c.Query("camera_id")

	cameraURL := "0" // Default to webcam

	// Priority 1: Specific Camera ID
	if camID != "" {
		var cam models.StationCamera
		if err := s.DB.First(&cam, camID).Error; err == nil {
			if cam.RTSPURL != "" {
				cameraURL = cam.RTSPURL
			}
		}
	} else if scaleID != "" {
		// Priority 2: Fallback to first camera of station (Legacy/Default)
		var station models.WeighingStation
		if err := s.DB.Preload("Cameras").First(&station, scaleID).Error; err == nil {
			if len(station.Cameras) > 0 {
				cameraURL = station.Cameras[0].RTSPURL
			} else if station.CameraURL != "" {
				cameraURL = station.CameraURL
			}
		}
	}

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
	session := sessions.Default(c)
	uidVal := session.Get("user_id")
	role := session.Get("role")

	// Filter IDs
	allowedIDs := make(map[uint]bool)
	if role == "admin" {
		// all allowed
	} else if uidVal != nil {
		var assignments []models.UserStationAssignment
		s.DB.Where("user_id = ?", uidVal).Find(&assignments)
		for _, a := range assignments {
			allowedIDs[a.WeighingStationID] = true
		}
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	c.Stream(func(w io.Writer) bool {
		if _, ok := <-ticker.C; ok {
			s.ScaleMgr.Mu.Lock()
			for id, scale := range s.ScaleMgr.Scales {
				// Only send data if allowed
				if role == "admin" || allowedIDs[id] {
					c.SSEvent("message", gin.H{
						"scale_id": id,
						"weight":   scale.LastWeight,
						"connected": scale.Connected,
					})
				}
			}
			s.ScaleMgr.Mu.Unlock()
			return true
		}
		return false
	})
}
