package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"stoneweigh/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestShowReports_BOLA(t *testing.T) {
	r, server := setupTestServer(t)

	// Add missing migrations and templates for this test
	assert.NoError(t, server.DB.AutoMigrate(&models.Company{}))
	r.SetHTMLTemplate(template.Must(template.New("reports.html").Parse("Records:{{range .Records}}{{.PlateNumber}},{{end}} Total:{{.TotalNetWeight}}")))

	r.GET("/reports", server.ShowReports)

	db := server.DB

	// Setup Data
	s1 := models.WeighingStation{Name: "Station 1", Enabled: true}
	s2 := models.WeighingStation{Name: "Station 2", Enabled: true}
	assert.NoError(t, db.Create(&s1).Error)
	assert.NoError(t, db.Create(&s2).Error)

	u1 := models.User{Username: "op1", Role: "operator"}
	assert.NoError(t, db.Create(&u1).Error)

	// Assign op1 to Station 1 ONLY
	assert.NoError(t, db.Create(&models.UserStationAssignment{UserID: u1.ID, WeighingStationID: s1.ID}).Error)

	now := time.Now()
	assert.NoError(t, db.Create(&models.WeighingRecord{
		PlateNumber:  "S1-CAR",
		ScaleID:      s1.ID,
		NetWeight:    100,
		GrossWeight:  1000,
		DriverName:   "Test Driver",
		WeighedAt:    now,
		TicketNumber: "T1",
	}).Error)
	assert.NoError(t, db.Create(&models.WeighingRecord{
		PlateNumber:  "S2-CAR",
		ScaleID:      s2.ID,
		NetWeight:    200,
		GrossWeight:  2000,
		DriverName:   "Test Driver",
		WeighedAt:    now,
		TicketNumber: "T2",
	}).Error)

	// Helper to login as u1
	r.POST("/login_op1", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("user_id", u1.ID)
		session.Set("role", u1.Role)
		session.Save()
		c.Status(200)
	})

	wLogin := httptest.NewRecorder()
	reqLogin, _ := http.NewRequest("POST", "/login_op1", nil)
	r.ServeHTTP(wLogin, reqLogin)
	cookie := wLogin.Header().Get("Set-Cookie")

	// Test: op1 should only see S1-CAR and total weight 100
	w := httptest.NewRecorder()
	url := "/reports?start_date=2000-01-01&end_date=2099-12-31"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Cookie", cookie)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	assert.Contains(t, body, "S1-CAR")
	assert.NotContains(t, body, "S2-CAR")
	assert.Contains(t, body, "Total:100")
}

func TestUpdateUserAssignments_UniqueConstraint(t *testing.T) {
	r, server := setupTestServer(t)
	r.POST("/api/users/:id/assignments", server.UpdateUserAssignments)

	db := server.DB

	u := models.User{Username: "testuser"}
	db.Create(&u)
	s := models.WeighingStation{Name: "Station X"}
	db.Create(&s)

	userIDStr := fmt.Sprintf("%d", u.ID)

	// 1. Assign once
	payload := map[string]any{"station_ids": []uint{s.ID}}
	body, _ := json.Marshal(payload)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/users/"+userIDStr+"/assignments", strings.NewReader(string(body)))
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 2. Clear assignments (causes soft delete if using Delete())
	payloadEmpty := map[string]any{"station_ids": []uint{}}
	bodyEmpty, _ := json.Marshal(payloadEmpty)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/users/"+userIDStr+"/assignments", strings.NewReader(string(bodyEmpty)))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// 3. Assign again (re-inserts record with same user/station)
	// If fix is not applied, this would fail due to unique constraint if record still exists (soft-deleted)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/api/users/"+userIDStr+"/assignments", strings.NewReader(string(body)))
	req3.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code, "Should be able to re-assign same station after clearing")
}
