package handlers

import (
	"net/http"
	"strconv"
	"stoneweigh/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// === User Management ===

func (s *Server) ShowUsers(c *gin.Context) {
	session := sessions.Default(c)
	fullName := "Operator"
	if v := session.Get("full_name"); v != nil {
		fullName = v.(string)
	}

	c.HTML(http.StatusOK, "users.html", gin.H{
		"title":       "User Management",
		"active":      "settings",
		"showNav":     true,
		"CurrentUser": fullName,
	})
}

// GetUsers returns all users with their assignments
func (s *Server) GetUsers(c *gin.Context) {
	var users []models.User
	if err := s.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	// Sanitize passwords
	for i := range users {
		users[i].PasswordHash = ""
	}
	c.JSON(http.StatusOK, users)
}

func (s *Server) CreateUser(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
		FullName string `json:"full_name"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username:     input.Username,
		PasswordHash: string(hash),
		FullName:     input.FullName,
		Role:         input.Role,
	}

	if err := s.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created"})
}

func (s *Server) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := s.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	// Also delete assignments
	s.DB.Where("user_id = ?", id).Delete(&models.UserStationAssignment{})

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

// === Assignments ===

func (s *Server) GetUserAssignments(c *gin.Context) {
	userID := c.Param("id")
	var assignments []models.UserStationAssignment
	if err := s.DB.Preload("WeighingStation").Where("user_id = ?", userID).Find(&assignments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
		return
	}
	c.JSON(http.StatusOK, assignments)
}

func (s *Server) UpdateUserAssignments(c *gin.Context) {
	userID := c.Param("id")
	var input struct {
		StationIDs []uint `json:"station_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := s.DB.Begin()
	// Clear existing
	if err := tx.Where("user_id = ?", userID).Delete(&models.UserStationAssignment{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update assignments"})
		return
	}

	// Add new
	for _, sid := range input.StationIDs {
		assign := models.UserStationAssignment{
			UserID:            stringToUint(userID),
			WeighingStationID: sid,
		}
		if err := tx.Create(&assign).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add assignment"})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Assignments updated"})
}

func stringToUint(s string) uint {
	val, _ := strconv.Atoi(s)
	return uint(val)
}
