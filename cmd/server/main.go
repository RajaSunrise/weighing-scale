package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"stoneweigh/internal/cv"
	"stoneweigh/internal/database"
	"stoneweigh/internal/handlers"
	"stoneweigh/internal/hardware"
	"stoneweigh/internal/models"
	"stoneweigh/internal/pkg/logger"
	"stoneweigh/internal/router"
)

func main() {
	// Set timezone to Asia/Jakarta
	if tz, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		time.Local = tz
		log.Println("Timezone set to Asia/Jakarta")
	}

	// Initialize Logger
	logger.Init()

	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 1. Initialize Database
	// We use the internal/database package which now handles the Postgres/Sqlite logic + migration
	database.Connect()
	db := database.DB

	// Seed Admin User
	seedAdmin(db)

	// 2. Initialize Hardware (Scales)
	hardware.InitScaleManager()

	// Load configs from DB
	// If no configs exist, maybe seed default ones for MVP (optional)
	// But `ReloadConfig` will handle empty lists gracefully.
	hardware.Manager.ReloadConfig(db)

	if os.Getenv("ENABLE_DEMO_SCALE") == "true" {
		hardware.Manager.StartDemoMode()
	}

	// 3. Initialize CV
	anpr := cv.NewANPRService("models/platdetection.onnx")

	// 4. Initialize Handlers
	server := handlers.NewServer(db, hardware.Manager, anpr)

	// 5. Setup Router
	r := router.SetupRouter(server)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("StoneWeigh Server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Server failed:", err)
	}
}

func seedAdmin(db *gorm.DB) {
	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")

	if username == "" || password == "" {
		return
	}

	var count int64
	db.Model(&models.User{}).Where("username = ?", username).Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		admin := models.User{
			Username:     username,
			PasswordHash: string(hash),
			FullName:     os.Getenv("ADMIN_FULLNAME"),
			Role:         "admin",
		}
		if err := db.Create(&admin).Error; err != nil {
			log.Printf("Failed to seed admin: %v", err)
		} else {
			log.Printf("Admin user '%s' seeded successfully", username)
		}
	}
}
