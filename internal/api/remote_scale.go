package api

import (
	"net/http"
	"stoneweigh/internal/database"
	"stoneweigh/internal/hardware"
	"stoneweigh/internal/models"

	"github.com/gin-gonic/gin"
)

type RemoteScalePayload struct {
	Weight float64 `json:"weight"`
}

// HandleRemoteScaleData receives weight data from a remote client
func HandleRemoteScaleData(c *gin.Context) {
	// 1. Get Token from Header (preferred)
	// SECURITY: Do not accept tokens via query parameters (leaks in logs/history)
	token := c.GetHeader("X-Scale-Token")

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

	// SECURITY: Range validation for weight to prevent data corruption or DoS
	if payload.Weight < 0 || payload.Weight > 1000000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Weight out of range (0-1,000,000)"})
		return
	}

	// 4. Update ScaleManager
	if hardware.Manager != nil {
		hardware.Manager.UpdateScale(station.ID, payload.Weight, true)
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
