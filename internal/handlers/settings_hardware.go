package handlers

import (
	"fmt"
	"net/http"
	"stoneweigh/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
	"gorm.io/gorm"
)

// === Weighing Station / Hardware Config ===

func (s *Server) GetStations(c *gin.Context) {
	var stations []models.WeighingStation
	if err := s.DB.Preload("Cameras").Find(&stations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stations"})
		return
	}
	c.JSON(http.StatusOK, stations)
}

func (s *Server) CreateStation(c *gin.Context) {
	var input models.WeighingStation
	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Printf("CreateStation JSON Bind Error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateStationInput(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create station"})
		return
	}

	// Reload hardware manager to apply changes
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

	if err := validateStationInput(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// SECURITY: Wrap updates in a transaction to ensure atomicity
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		// Update fields
		station.Name = input.Name
		station.ScalePort = input.ScalePort
		station.BaudRate = input.BaudRate
		station.Enabled = input.Enabled
		station.Token = input.Token

		// Handle Cameras update
		// 1. Delete existing cameras
		if err := tx.Where("weighing_station_id = ?", station.ID).Delete(&models.StationCamera{}).Error; err != nil {
			return err
		}
		// 2. Add new ones
		station.Cameras = input.Cameras

		if err := tx.Save(&station).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
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
		"csrf_token":  csrf.GetToken(c),
	})
}

func validateStationInput(input *models.WeighingStation) error {
	if err := validateLength(input.Name, "Name", 1, 100); err != nil {
		return err
	}
	if err := validateSafeString(input.Name, "Name"); err != nil {
		return err
	}

	if err := validateLength(input.ScalePort, "ScalePort", 1, 50); err != nil {
		return err
	}
	if err := validateSafeString(input.ScalePort, "ScalePort"); err != nil {
		return err
	}

	if input.Token != nil && *input.Token != "" {
		if err := validateLength(*input.Token, "Token", 0, 64); err != nil {
			return err
		}
		if err := validateSafeString(*input.Token, "Token"); err != nil {
			return err
		}
	}

	for _, cam := range input.Cameras {
		if err := validateLength(cam.Name, "Camera Name", 1, 50); err != nil {
			return err
		}
		if err := validateSafeString(cam.Name, "Camera Name"); err != nil {
			return err
		}
		if err := validateLength(cam.RTSPURL, "RTSP URL", 0, 255); err != nil {
			return err
		}
		if err := validateSafeString(cam.RTSPURL, "RTSP URL"); err != nil {
			return err
		}
	}

	return nil
}
