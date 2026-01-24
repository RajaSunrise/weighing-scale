package handlers

import (
	"encoding/json"
	"html/template"
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

// setupServer sets up a test server with in-memory DB and mocks
func setupServer(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	// Setup In-Memory DB
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate Models
	err = db.AutoMigrate(
		&models.User{},
		&models.WeighingStation{},
		&models.StationCamera{},
		&models.UserStationAssignment{},
		&models.WeighingRecord{},
	)
	assert.NoError(t, err)

	// Mock Services
	// Initialize ScaleManager global manually or use local if possible.
	// hardware.InitScaleManager() usually sets global.
	// But NewServer takes ScaleManager.
	hardware.InitScaleManager()
	scaleMgr := hardware.Manager

	// ANPR Service Mock (using mock implementation via build tags or forced path)
	// Since we are running tests, usually mock file is active if build tags align,
	// OR we can just instantiate it.
	// The mock version of NewANPRService takes 1 arg.
	anprService := cv.NewANPRService("mock_model.onnx")

	server := NewServer(db, scaleMgr, anprService, nil)

	// Setup Router with Session Middleware (needed for Handlers)
	r := gin.New()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// Register Routes
	r.POST("/api/transaction", server.SaveTransaction)
	r.POST("/api/anpr/trigger", server.TriggerANPR)
	r.GET("/dashboard", server.ShowDashboard)

	return r, db
}

func TestSaveTransaction(t *testing.T) {
	r, db := setupServer(t)

	// Helper to set session
	r.POST("/login_mock", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("username", "TestManager")
		session.Save()
		c.Status(200)
	})

	// 1. Establish Session
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login_mock", nil)
	r.ServeHTTP(w, req)
	cookie := w.Header().Get("Set-Cookie")

	// 2. Send Valid Transaction
	payload := `{"scale_id": 1, "plate_number": "B 1234 XX", "driver_name": "John", "company": "ABC Corp", "product": "Stone", "gross": 15000, "tare": 5000}`
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/transaction", strings.NewReader(payload))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Cookie", cookie)
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Transaction saved", response["message"])
	assert.NotEmpty(t, response["ticket"])
	assert.NotEmpty(t, response["invoice"])

	// 3. Verify DB
	var record models.WeighingRecord
	err = db.First(&record).Error
	assert.NoError(t, err)
	assert.Equal(t, "B 1234 XX", record.PlateNumber)
	assert.Equal(t, 10000.0, record.NetWeight)
	assert.Equal(t, "TestManager", record.ManagerName)
}

func TestSaveTransaction_InvalidJSON(t *testing.T) {
	r, _ := setupServer(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/transaction", strings.NewReader(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTriggerANPR(t *testing.T) {
	r, db := setupServer(t)

	// Setup Data: Station with Camera
	station := models.WeighingStation{Name: "Station 1", Enabled: true}
	db.Create(&station)
	camera := models.StationCamera{Name: "Cam 1", RTSPURL: "rtsp://mock", WeighingStationID: station.ID}
	db.Create(&camera)

	// Test 1: By Camera ID (Success Mock)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/anpr/trigger?camera_id=1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Contains(t, []string{"success", "simulated"}, resp["status"])
}

func TestDashboardStats(t *testing.T) {
	r, db := setupServer(t)

	// Mock Template
	t.Setenv("SESSION_SECRET", "test")
	tmpl := template.New("")
	tmpl.Funcs(template.FuncMap{
		"dict": func(v ...interface{}) (map[string]interface{}, error) { return nil, nil },
		"json": func(v any) template.JS { return "" },
		"currentYear": func() int { return 2026 },
	})
	tmpl.Parse(`{{define "dashboard.html"}}Dashboard: Count={{.Stats.TodayCount}}, Weight={{.Stats.TodayWeight}}{{end}}`)
	r.SetHTMLTemplate(tmpl)

	// Seed Data
	now := time.Now()

	// r1 and r2 are "today"
	r1 := models.WeighingRecord{
		TicketNumber: "T1", PlateNumber: "B 1111", DriverName: "D1", GrossWeight: 200, TareWeight: 100,
		NetWeight: 100, WeighedAt: now,
	}
	r2 := models.WeighingRecord{
		TicketNumber: "T2", PlateNumber: "B 2222", DriverName: "D2", GrossWeight: 300, TareWeight: 100,
		NetWeight: 200, WeighedAt: now.Add(1 * time.Minute),
	}
	// r3 is "2 days ago"
	r3 := models.WeighingRecord{
		TicketNumber: "T3", PlateNumber: "B 3333", DriverName: "D3", GrossWeight: 400, TareWeight: 100,
		NetWeight: 300, WeighedAt: now.Add(-48 * time.Hour),
	}

	assert.NoError(t, db.Create(&r1).Error)
	assert.NoError(t, db.Create(&r2).Error)
	assert.NoError(t, db.Create(&r3).Error)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Assert strict values: Count should be 2, Weight should be 300 (100+200)
	assert.Contains(t, w.Body.String(), "Dashboard: Count=2, Weight=300")
}
