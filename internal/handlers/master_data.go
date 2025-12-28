package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"stoneweigh/internal/models"
)

// ShowVehicleSettings renders the vehicle management page
func (s *Server) ShowVehicleSettings(c *gin.Context) {
	session := sessions.Default(c)
	fullName := "Operator"
	if v := session.Get("full_name"); v != nil {
		fullName = v.(string)
	}

	c.HTML(http.StatusOK, "settings_vehicles.html", gin.H{
		"title":       "Vehicle Management",
		"active":      "settings",
		"showNav":     true,
		"CurrentUser": fullName,
	})
}

// ListVehicles API returns all registered vehicles
func (s *Server) ListVehicles(c *gin.Context) {
	var vehicles []models.Vehicle
	if err := s.DB.Find(&vehicles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vehicles"})
		return
	}
	c.JSON(http.StatusOK, vehicles)
}

// CreateVehicle API adds a new vehicle
func (s *Server) CreateVehicle(c *gin.Context) {
	var input struct {
		PlateNumber  string  `json:"plate_number" binding:"required"`
		DriverName   string  `json:"driver_name" binding:"required"`
		DefaultTare  float64 `json:"default_tare"`
		OwnerCompany string  `json:"owner_company"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vehicle := models.Vehicle{
		PlateNumber:  input.PlateNumber,
		DriverName:   input.DriverName,
		DefaultTare:  input.DefaultTare,
		OwnerCompany: input.OwnerCompany,
	}

	if err := s.DB.Create(&vehicle).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vehicle. Plate number might be duplicate."})
		return
	}

	c.JSON(http.StatusCreated, vehicle)
}

// DeleteVehicle API removes a vehicle
func (s *Server) DeleteVehicle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := s.DB.Delete(&models.Vehicle{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vehicle"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vehicle deleted"})
}

// GetVehicleDetails returns details for a specific plate (public for operators)
func (s *Server) GetVehicleDetails(c *gin.Context) {
	plate := c.Query("plate")
	if plate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plate number required"})
		return
	}

	var vehicle models.Vehicle
	if err := s.DB.Where("plate_number = ?", plate).First(&vehicle).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vehicle not found"})
		return
	}

	c.JSON(http.StatusOK, vehicle)
}

// SearchVehicles performs a fuzzy search for autocomplete
func (s *Server) SearchVehicles(c *gin.Context) {
	query := strings.ToUpper(strings.TrimSpace(c.Query("q")))
	log.Printf("SearchVehicles called with query: '%s' (length: %d)", query, len(query))

	// Return empty array for empty query to prevent returning all vehicles
	if len(query) < 1 {
		c.JSON(http.StatusOK, []models.Vehicle{})
		return
	}

	var vehicles []models.Vehicle
	// Simple fuzzy search - case insensitive (already uppercased)
	err := s.DB.Where("plate_number LIKE ?", "%"+query+"%").Limit(10).Find(&vehicles).Error
	if err != nil {
		log.Printf("SearchVehicles error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	log.Printf("SearchVehicles found %d vehicles", len(vehicles))
	c.JSON(http.StatusOK, vehicles)
}
