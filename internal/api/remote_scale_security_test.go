package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stoneweigh/internal/database"
	"stoneweigh/internal/hardware"
	"stoneweigh/internal/models"
)

func TestHandleRemoteScaleData_Security_QueryParam(t *testing.T) {
	// 1. Setup DB
	dsn := fmt.Sprintf("file:memdb_sec_%d?mode=memory&cache=shared", 1)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	db.AutoMigrate(&models.WeighingStation{})
	database.DB = db

	// 2. Setup Hardware Manager
	hardware.InitScaleManager()

	// 3. Seed Station
	token := "secure_token_123"
	station := models.WeighingStation{
		Name:    "Secure Station",
		Enabled: true,
		Token:   &token,
	}
	assert.NoError(t, db.Create(&station).Error)

	// Register in manager
	hardware.Manager.Mu.Lock()
	hardware.Manager.Scales[station.ID] = &hardware.ScaleConnection{
		Config: station,
	}
	hardware.Manager.Mu.Unlock()

	// 4. Setup Router
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/external/scale", HandleRemoteScaleData)

	payload := map[string]interface{}{"weight": 500.0}
	body, _ := json.Marshal(payload)

	// Test Case 1: Query Param (Should Fail after fix, but let's see if we can catch it passing first?)
	// Actually, I am writing this test to assert the DESIRED behavior (Fail).
	// So if I run this BEFORE the fix, it should FAIL the assertion (because it will return 200 OK).

	t.Run("Reject Query Parameter Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/external/scale?token="+token, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		r.ServeHTTP(w, req)

		// Expect 401 Unauthorized because we don't want to accept query params
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Accept Header Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/external/scale", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Scale-Token", token)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
