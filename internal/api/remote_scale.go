package api

import (
	"net/http"
	"stoneweigh/internal/database"
	"stoneweigh/internal/hardware"
	"stoneweigh/internal/models"
	"time"

	"github.com/gin-gonic/gin"
)

type RemoteScalePayload struct {
	Weight float64 `json:"weight"`
}

// HandleRemoteScaleData receives weight data from a remote client
func HandleRemoteScaleData(c *gin.Context) {
	// 1. Get Token from Header (preferred) or Query
	token := c.GetHeader("X-Scale-Token")
	if token == "" {
		token = c.Query("token")
	}

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token authentication required"})
		return
	}

	// 2. Validate Token and find Station
	var station models.WeighingStation
	if err := database.DB.Where("token = ? AND enabled = ?", token, true).First(&station).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or inactive token"})
		return
	}

	// 3. Parse Body
	var payload RemoteScalePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// 4. Broadcast to ScaleManager
	// We inject this directly into the DataChannel which the SSE handler listens to.
	if hardware.Manager != nil {
		hardware.Manager.DataChannel <- hardware.ScaleData{
			ScaleID:   station.ID,
			Weight:    payload.Weight,
			Connected: true,
			Timestamp: time.Now().Unix(),
		}
	} else {
		// Should not happen if server is running correctly
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Scale manager not initialized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"station": station.Name,
		"received_weight": payload.Weight,
	})
}
