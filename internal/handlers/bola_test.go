package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"stoneweigh/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDashboardBOLA(t *testing.T) {
	r, db := setupServer(t)

	// Mock Template
	tmpl := template.New("")
	tmpl.Funcs(template.FuncMap{
		"dict":        func(v ...any) (map[string]any, error) { return nil, nil },
		"json":        func(v any) template.JS { return "" },
		"currentYear": func() int { return 2026 },
	})
	tmpl.Parse(`{{define "dashboard.html"}}Count={{.Stats.TodayCount}} Weight={{.Stats.TodayWeight}} Recent={{len .Recent}}{{end}}`)
	r.SetHTMLTemplate(tmpl)

	// 1. Setup Stations
	s1 := models.WeighingStation{Name: "Station 1", Enabled: true}
	s2 := models.WeighingStation{Name: "Station 2", Enabled: true}
	db.Create(&s1)
	db.Create(&s2)

	// 2. Setup Records
	now := time.Now()
	// Record for Station 1
	assert.NoError(t, db.Create(&models.WeighingRecord{
		TicketNumber: "T1", ScaleID: s1.ID, NetWeight: 100, GrossWeight: 1000, WeighedAt: now, PlateNumber: "B1", DriverName: "D1",
	}).Error)
	// Record for Station 2
	assert.NoError(t, db.Create(&models.WeighingRecord{
		TicketNumber: "T2", ScaleID: s2.ID, NetWeight: 200, GrossWeight: 2000, WeighedAt: now, PlateNumber: "B2", DriverName: "D2",
	}).Error)

	// 3. Setup Operator assigned ONLY to Station 1
	op := models.User{Username: "op1", Role: "operator", FullName: "Operator 1"}
	assert.NoError(t, db.Create(&op).Error)
	assert.NoError(t, db.Create(&models.UserStationAssignment{UserID: op.ID, WeighingStationID: s1.ID}).Error)

	// Helper to login
	r.POST("/login_op", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("user_id", op.ID)
		session.Set("role", "operator")
		session.Save()
		c.Status(200)
	})

	wLogin := httptest.NewRecorder()
	reqLogin, _ := http.NewRequest("POST", "/login_op", nil)
	r.ServeHTTP(wLogin, reqLogin)
	cookie := wLogin.Header().Get("Set-Cookie")

	// 4. Test Dashboard as Operator
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/dashboard", nil)
	req.Header.Set("Cookie", cookie)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Should ONLY see Station 1 record (Count=1, Weight=100, Recent=1)
	assert.Contains(t, w.Body.String(), "Count=1 Weight=100 Recent=1")

	// 5. Test Dashboard as Admin (should see everything)
	r.POST("/login_admin_mock", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("user_id", uint(999))
		session.Set("role", "admin")
		session.Save()
		c.Status(200)
	})
	wLoginAdmin := httptest.NewRecorder()
	reqLoginAdmin, _ := http.NewRequest("POST", "/login_admin_mock", nil)
	r.ServeHTTP(wLoginAdmin, reqLoginAdmin)
	adminCookie := wLoginAdmin.Header().Get("Set-Cookie")

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/dashboard", nil)
	req2.Header.Set("Cookie", adminCookie)
	r.ServeHTTP(w2, req2)
	assert.Contains(t, w2.Body.String(), "Count=2 Weight=300 Recent=2")
}

func TestReportChartsBOLA(t *testing.T) {
	r, db := setupServer(t)

	// 1. Setup Stations
	s1 := models.WeighingStation{Name: "Station 1", Enabled: true}
	s2 := models.WeighingStation{Name: "Station 2", Enabled: true}
	db.Create(&s1)
	db.Create(&s2)

	// 2. Setup Records
	now := time.Now()
	// Record for Station 1
	assert.NoError(t, db.Create(&models.WeighingRecord{
		TicketNumber: "T1", ScaleID: s1.ID, NetWeight: 500, GrossWeight: 1000, WeighedAt: now, PlateNumber: "B1", DriverName: "D1",
	}).Error)
	// Record for Station 2
	assert.NoError(t, db.Create(&models.WeighingRecord{
		TicketNumber: "T2", ScaleID: s2.ID, NetWeight: 1000, GrossWeight: 2000, WeighedAt: now, PlateNumber: "B2", DriverName: "D2",
	}).Error)

	// 3. Setup Operator assigned ONLY to Station 1
	op := models.User{Username: "op2", Role: "operator", FullName: "Operator 2"}
	assert.NoError(t, db.Create(&op).Error)
	assert.NoError(t, db.Create(&models.UserStationAssignment{UserID: op.ID, WeighingStationID: s1.ID}).Error)

	// Login helper
	r.POST("/login_op2", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("user_id", op.ID)
		session.Set("role", "operator")
		session.Save()
		c.Status(200)
	})
	wLogin := httptest.NewRecorder()
	reqLogin, _ := http.NewRequest("POST", "/login_op2", nil)
	r.ServeHTTP(wLogin, reqLogin)
	cookie := wLogin.Header().Get("Set-Cookie")

	// 4. Test Charts as Operator
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/reports/charts?period=daily", nil)
	req.Header.Set("Cookie", cookie)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp ChartData
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Find today's data in the chart
	todayLabel := now.Format("02 Jan")
	found := false
	for i, label := range resp.Labels {
		if label == todayLabel {
			assert.Equal(t, 500.0, resp.Data[i], "Operator should only see 500kg from Station 1")
			found = true
		}
	}
	assert.True(t, found, "Today's label not found in chart")

	// 5. Test Charts as Admin
	r.POST("/login_admin_mock2", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("user_id", uint(998))
		session.Set("role", "admin")
		session.Save()
		c.Status(200)
	})
	wLoginAdmin := httptest.NewRecorder()
	reqLoginAdmin, _ := http.NewRequest("POST", "/login_admin_mock2", nil)
	r.ServeHTTP(wLoginAdmin, reqLoginAdmin)
	adminCookie := wLoginAdmin.Header().Get("Set-Cookie")

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/reports/charts?period=daily", nil)
	req2.Header.Set("Cookie", adminCookie)
	r.ServeHTTP(w2, req2)
	var resp2 ChartData
	json.Unmarshal(w2.Body.Bytes(), &resp2)

	for i, label := range resp2.Labels {
		if label == todayLabel {
			assert.Equal(t, 1500.0, resp2.Data[i], "Admin should see 1500kg (500+1000)")
		}
	}
}
