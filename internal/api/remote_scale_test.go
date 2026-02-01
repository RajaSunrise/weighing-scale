package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stoneweigh/internal/database"
	"stoneweigh/internal/hardware"
	"stoneweigh/internal/models"
)

func TestHandleRemoteScaleData(t *testing.T) {
	// 1. Setup DB
	dsn := "file:memdb_api?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&models.WeighingStation{})
	database.DB = db // Set global DB

	// 2. Setup Hardware Manager
	hardware.InitScaleManager() // Should not create blocking channel anymore

	// 3. Seed Station
	token := "test_token"
	station := models.WeighingStation{
		Name:    "Test Station",
		Enabled: true,
		Token:   &token,
	}
	assert.NoError(t, db.Create(&station).Error)

	// Manually add to Manager to avoid starting monitor goroutines
	hardware.Manager.Mu.Lock()
	hardware.Manager.Scales[station.ID] = &hardware.ScaleConnection{
		Config: station,
	}
	hardware.Manager.Mu.Unlock()

	// 4. Setup Router
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/external/scale", HandleRemoteScaleData)

	// 5. Perform Request
	payload := map[string]interface{}{
		"weight": 123.45,
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/external/scale", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Scale-Token", token)

	// Measure time to ensure no blocking
	start := time.Now()
	r.ServeHTTP(w, req)
	duration := time.Since(start)

	// 6. Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Less(t, duration, 100*time.Millisecond, "Request took too long, potential blocking")

	// Verify ScaleManager state updated
	hardware.Manager.Mu.Lock()
	scale, exists := hardware.Manager.Scales[station.ID]
	hardware.Manager.Mu.Unlock()

	assert.True(t, exists)
	assert.Equal(t, 123.45, scale.GetWeight())
	assert.True(t, scale.IsConnected())
}
