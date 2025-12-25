package handlers

import (
	"net/http"
	"stoneweigh/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// === Weighing Station / Hardware Config ===

func (s *Server) GetStations(c *gin.Context) {
	var stations []models.WeighingStation
	if err := s.DB.Find(&stations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stations"})
		return
	}
	c.JSON(http.StatusOK, stations)
}

func (s *Server) CreateStation(c *gin.Context) {
	var input models.WeighingStation
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create station"})
		return
	}

	// Reload hardware manager to apply changes
	// Note: In a real distributed system, we'd need a pub/sub.
	// For this single instance, we can call directly.
	go s.ScaleMgr.ReloadConfig(s.DB)

	c.JSON(http.StatusOK, input)
}

func (s *Server) UpdateStation(c *gin.Context) {
	id := c.Param("id")
	var station models.WeighingStation
	if err := s.DB.First(&station, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Station not found"})
		return
	}

	var input models.WeighingStation
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	station.Name = input.Name
	station.ScalePort = input.ScalePort
	station.BaudRate = input.BaudRate
	station.CameraURL = input.CameraURL
	station.Enabled = input.Enabled

	if err := s.DB.Save(&station).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update station"})
		return
	}

	go s.ScaleMgr.ReloadConfig(s.DB)

	c.JSON(http.StatusOK, station)
}

func (s *Server) DeleteStation(c *gin.Context) {
	id := c.Param("id")
	if err := s.DB.Delete(&models.WeighingStation{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete station"})
		return
	}

	go s.ScaleMgr.ReloadConfig(s.DB)

	c.JSON(http.StatusOK, gin.H{"message": "Station deleted"})
}

// ShowSettings renders the settings page
// We overload this to show the "Hardware" tab data if needed, or rely on AJAX.
// The existing `settings.html` seems to be for general settings.
// We will modify it to act as a container or add a specific route.
func (s *Server) ShowSettingsHardware(c *gin.Context) {
	session := sessions.Default(c)
	fullName := "Operator"
	if v := session.Get("full_name"); v != nil {
		fullName = v.(string)
	}

	c.HTML(http.StatusOK, "settings_hardware.html", gin.H{
		"title":       "Hardware Settings",
		"active":      "settings",
		"showNav":     true,
		"CurrentUser": fullName,
	})
}
