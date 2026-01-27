package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"stoneweigh/internal/cv"
	"stoneweigh/internal/hardware"
	"stoneweigh/internal/models"
	"stoneweigh/internal/pkg"
	"stoneweigh/internal/reporting"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	csrf "github.com/utrack/gin-csrf"
	"gorm.io/gorm"
)

type Server struct {
	DB          *gorm.DB
	ScaleMgr    *hardware.ScaleManager
	ANPRService *cv.ANPRService
	Redis       *redis.Client
}

func NewServer(db *gorm.DB, sm *hardware.ScaleManager, anpr *cv.ANPRService, rdb *redis.Client) *Server {
	return &Server{DB: db, ScaleMgr: sm, ANPRService: anpr, Redis: rdb}
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

	cacheKey := "dashboard:stats:today"
	cached := false

	// Try Cache
	if s.Redis != nil {
		val, err := s.Redis.Get(c, cacheKey).Result()
		if err == nil {
			var cachedStats struct {
				Count int64
				Total float64
			}
			if err := json.Unmarshal([]byte(val), &cachedStats); err == nil {
				todayCount = cachedStats.Count
				todayWeight = cachedStats.Total
				cached = true
			}
		}
	}

	if !cached {
		// Optimization: Fetch Count and Sum in one query
		type StatsResult struct {
			Count int64
			Total float64
		}
		var stats StatsResult

		s.DB.Model(&models.WeighingRecord{}).
			Select("count(*) as count, COALESCE(sum(net_weight), 0) as total").
			Where("weighed_at >= ?", startOfDay).
			Scan(&stats)

		todayCount = stats.Count
		todayWeight = stats.Total

		// Save to Cache (5 minutes)
		if s.Redis != nil {
			data, _ := json.Marshal(stats)
			s.Redis.Set(c, cacheKey, data, 5*time.Minute)
		}
	}

	// 2. Fetch Recent Transactions
	var recent []models.WeighingRecord

	recentCacheKey := "dashboard:recent"
	recentCached := false

	if s.Redis != nil {
		val, err := s.Redis.Get(c, recentCacheKey).Result()
		if err == nil {
			if err := json.Unmarshal([]byte(val), &recent); err == nil {
				recentCached = true
			}
		}
	}

	if !recentCached {
		s.DB.Order("weighed_at desc").Limit(10).Find(&recent)
		if s.Redis != nil {
			data, _ := json.Marshal(recent)
			s.Redis.Set(c, recentCacheKey, data, 5*time.Minute)
		}
	}

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
		"csrf_token":  csrf.GetToken(c),
	})
}

// ShowReports renders the report page
func (s *Server) ShowReports(c *gin.Context) {
	session := sessions.Default(c)
	fullName := "Operator"
	if v := session.Get("full_name"); v != nil {
		fullName = v.(string)
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	companyFilter := c.Query("company") // Expecting company name or ID

	// Default to today if empty
	if startDate == "" {
		startDate = time.Now().Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	// Parse
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)
	// Set end to end of day
	end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// Define filter scope to reuse in both queries
	filterScope := func(db *gorm.DB) *gorm.DB {
		db = db.Where("weighed_at BETWEEN ? AND ?", start, end)
		if companyFilter != "" {
			db = db.Where("company_name = ?", companyFilter)
		}
		return db
	}

	var records []models.WeighingRecord
	s.DB.Model(&models.WeighingRecord{}).Scopes(filterScope).
		Order("weighed_at desc").Find(&records)

	// Optimization: Calculate total in memory to save a DB query
	var totalNet float64
	for _, r := range records {
		totalNet += r.NetWeight
	}

	// Fetch distinct companies for filter dropdown
	var companies []models.Company
	s.DB.Order("name asc").Find(&companies)

	c.HTML(http.StatusOK, "reports.html", gin.H{
		"title":           "Reports",
		"active":          "reports",
		"showNav":         true,
		"CurrentUser":     fullName,
		"Records":         records,
		"StartDate":       startDate,
		"EndDate":         endDate,
		"TotalNetWeight":  totalNet,
		"Companies":       companies,
		"SelectedCompany": companyFilter,
		"PrintDate":       time.Now(),
		"csrf_token":      csrf.GetToken(c),
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

	log.Printf("Transaction Data - Plate: %s, Driver: %s, Company: %s, Product: %s, Gross: %.2f, Tare: %.2f",
		input.PlateNumber, input.DriverName, input.Company, input.Product, input.Gross, input.Tare)

	session := sessions.Default(c)
	managerName := "Unknown"
	if val := session.Get("username"); val != nil {
		managerName = val.(string)
	}

	// SECURITY CHECK: Ensure user is allowed to write to this station
	role := session.Get("role")
	userID := session.Get("user_id")

	if role != "admin" {
		var count int64
		s.DB.Model(&models.UserStationAssignment{}).
			Where("user_id = ? AND weighing_station_id = ?", userID, input.ScaleID).
			Count(&count)

		if count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this station"})
			return
		}
	}

	net := input.Gross - input.Tare
	// Use UnixNano to prevent collision on rapid submissions
	ticket, err := pkg.GenerateTicketID(12)
	if err != nil {
		panic(err)
	}

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

	// Invalidate Dashboard Cache
	if s.Redis != nil {
		s.Redis.Del(c, "dashboard:stats:today")
		s.Redis.Del(c, "dashboard:recent")
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
	var targetStationID uint

	// Priority 1: Specific Camera ID
	if camID != "" {
		var cam models.StationCamera
		if err := s.DB.First(&cam, camID).Error; err == nil {
			if cam.RTSPURL != "" {
				cameraURL = cam.RTSPURL
			}
			targetStationID = cam.WeighingStationID
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Camera not found"})
			return
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
			targetStationID = station.ID
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Station not found"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing station or camera ID"})
		return
	}

	// SECURITY CHECK: Ensure user is allowed to access this station's hardware
	session := sessions.Default(c)
	role := session.Get("role")
	userID := session.Get("user_id")

	if role != "admin" {
		var count int64
		s.DB.Model(&models.UserStationAssignment{}).
			Where("user_id = ? AND weighing_station_id = ?", userID, targetStationID).
			Count(&count)

		if count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this station hardware"})
			return
		}
	}

	plate, snapshotPath, err := s.ANPRService.CaptureAndDetect(cameraURL)
	if err != nil {
		// Fallback for demo/simulation if no camera
		c.JSON(http.StatusOK, gin.H{
			"plate":    "B 1234 DEMO",
			"snapshot": "/static/images/placeholder_truck.jpg",
			"status":   "simulated",
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
			type scaleSnapshot struct {
				ID        uint
				Weight    float64
				Connected bool
			}
			snapshots := make([]scaleSnapshot, 0, len(s.ScaleMgr.Scales))

			for id, scale := range s.ScaleMgr.Scales {
				// Only send data if allowed
				if role == "admin" || allowedIDs[id] {
					snapshots = append(snapshots, scaleSnapshot{
						ID:        id,
						Weight:    scale.LastWeight,
						Connected: scale.Connected,
					})
				}
			}
			s.ScaleMgr.Mu.Unlock()

			for _, snap := range snapshots {
				c.SSEvent("message", gin.H{
					"scale_id":  snap.ID,
					"weight":    snap.Weight,
					"connected": snap.Connected,
				})
			}
			return true
		}
		return false
	})
}
