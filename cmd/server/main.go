package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stoneweigh/internal/cv"
	"stoneweigh/internal/handlers"
	"stoneweigh/internal/hardware"
	"stoneweigh/internal/models"
	"stoneweigh/internal/router"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// 1. Initialize Database
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "stoneweigh.db"
	}
	db, err := gorm.Open(sqlite.Open(dbDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Migrate Schema
	db.AutoMigrate(&models.WeighingRecord{}, &models.ScaleConfig{}, &models.Vehicle{}, &models.Invoice{}, &models.User{})

	// Seed Admin User
	seedAdmin(db)

	// 2. Initialize Hardware (Scales)
	hardware.InitScaleManager()
	// Add default scales (Mocking ports for now - in prod use env vars to loop)
	hardware.Manager.AddScale(models.ScaleConfig{Model: gorm.Model{ID: 1}, Name: "Main Gate", Port: "COM3", BaudRate: 9600, Enabled: true})
	hardware.Manager.AddScale(models.ScaleConfig{Model: gorm.Model{ID: 2}, Name: "Side Gate", Port: "COM4", BaudRate: 9600, Enabled: true})
	hardware.Manager.AddScale(models.ScaleConfig{Model: gorm.Model{ID: 3}, Name: "Back Gate", Port: "COM5", BaudRate: 9600, Enabled: true})

	// 3. Initialize CV
	anpr := cv.NewANPRService("models/platdetection.pt")

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
