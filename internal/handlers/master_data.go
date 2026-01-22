package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
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
		"csrf_token":  csrf.GetToken(c),
	})
}

// ListVehicles API returns all registered vehicles
func (s *Server) ListVehicles(c *gin.Context) {
	var vehicles []models.Vehicle
	if err := s.DB.Preload("Company").Find(&vehicles).Error; err != nil {
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
		// Legacy string field, but we can fill it from Company name if ID provided
		OwnerCompany string  `json:"owner_company"`

		// New fields
		SIM          string  `json:"sim"`
		CompanyID    *uint   `json:"company_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If CompanyID is provided, fetch the company name to populate OwnerCompany for legacy/display compatibility
	if input.CompanyID != nil && *input.CompanyID > 0 {
		var comp models.Company
		if err := s.DB.First(&comp, *input.CompanyID).Error; err == nil {
			input.OwnerCompany = comp.Name
		}
	}

	vehicle := models.Vehicle{
		PlateNumber:  input.PlateNumber,
		DriverName:   input.DriverName,
		DefaultTare:  input.DefaultTare,
		OwnerCompany: input.OwnerCompany,
		SIM:          input.SIM,
		CompanyID:    input.CompanyID,
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

// UpdateVehicle API updates an existing vehicle
func (s *Server) UpdateVehicle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input struct {
		PlateNumber  string  `json:"plate_number" binding:"required"`
		DriverName   string  `json:"driver_name" binding:"required"`
		DefaultTare  float64 `json:"default_tare"`
		OwnerCompany string  `json:"owner_company"` // Legacy/Fallback

		// New fields
		SIM       string `json:"sim"`
		CompanyID *uint  `json:"company_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var vehicle models.Vehicle
	if err := s.DB.First(&vehicle, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vehicle not found"})
		return
	}

	// Update fields
	vehicle.PlateNumber = input.PlateNumber
	vehicle.DriverName = input.DriverName
	vehicle.DefaultTare = input.DefaultTare
	vehicle.SIM = input.SIM
	vehicle.CompanyID = input.CompanyID

	// If CompanyID is provided, fetch the company name to populate OwnerCompany
	if input.CompanyID != nil && *input.CompanyID > 0 {
		var comp models.Company
		if err := s.DB.First(&comp, *input.CompanyID).Error; err == nil {
			vehicle.OwnerCompany = comp.Name
		}
	} else {
		vehicle.OwnerCompany = input.OwnerCompany
	}

	if err := s.DB.Save(&vehicle).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vehicle"})
		return
	}

	c.JSON(http.StatusOK, vehicle)
}

// GetVehicleDetails returns details for a specific plate (public for operators)
func (s *Server) GetVehicleDetails(c *gin.Context) {
	plate := c.Query("plate")
	if plate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Plate number required"})
		return
	}

	var vehicle models.Vehicle
	if err := s.DB.Preload("Company").Where("plate_number = ?", plate).First(&vehicle).Error; err != nil {
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
	err := s.DB.Preload("Company").Where("plate_number LIKE ?", "%"+query+"%").Limit(10).Find(&vehicles).Error
	if err != nil {
		log.Printf("SearchVehicles error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	log.Printf("SearchVehicles found %d vehicles", len(vehicles))
	c.JSON(http.StatusOK, vehicles)
}

// === COMPANY MANAGEMENT ===

// ListCompanies API
func (s *Server) ListCompanies(c *gin.Context) {
	var companies []models.Company
	if err := s.DB.Order("name asc").Find(&companies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch companies"})
		return
	}
	c.JSON(http.StatusOK, companies)
}

// CreateCompany API
func (s *Server) CreateCompany(c *gin.Context) {
	var input struct {
		Name          string `json:"name" binding:"required"`
		Address       string `json:"address"`
		ContactPerson string `json:"contact_person"`
		Phone         string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	company := models.Company{
		Name:          input.Name,
		Address:       input.Address,
		ContactPerson: input.ContactPerson,
		Phone:         input.Phone,
	}

	if err := s.DB.Create(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company. Name might be duplicate."})
		return
	}

	c.JSON(http.StatusCreated, company)
}

// UpdateCompany API
func (s *Server) UpdateCompany(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var input struct {
		Name          string `json:"name" binding:"required"`
		Address       string `json:"address"`
		ContactPerson string `json:"contact_person"`
		Phone         string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var company models.Company
	if err := s.DB.First(&company, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}

	company.Name = input.Name
	company.Address = input.Address
	company.ContactPerson = input.ContactPerson
	company.Phone = input.Phone

	if err := s.DB.Save(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update company"})
		return
	}

	c.JSON(http.StatusOK, company)
}

// DeleteCompany API
func (s *Server) DeleteCompany(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Optional: Check if used by vehicles?
	// GORM foreign keys might restrict this depending on constraint setup.

	if err := s.DB.Delete(&models.Company{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete company"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Company deleted"})
}
