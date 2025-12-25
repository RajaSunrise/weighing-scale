package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"stoneweigh/internal/cv"
	"stoneweigh/internal/handlers"
	"stoneweigh/internal/hardware"
	"stoneweigh/internal/models"
)

func main() {
	// 1. Initialize Database
	db, err := gorm.Open(sqlite.Open("stoneweigh.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Migrate Schema
	db.AutoMigrate(&models.WeighingRecord{}, &models.ScaleConfig{}, &models.Vehicle{}, &models.Invoice{})

	// 2. Initialize Hardware (Scales)
	hardware.InitScaleManager()
	// Add default scales (Mocking ports for now)
	hardware.Manager.AddScale(models.ScaleConfig{Model: gorm.Model{ID: 1}, Name: "Main Gate", Port: "COM3", BaudRate: 9600, Enabled: true})
	hardware.Manager.AddScale(models.ScaleConfig{Model: gorm.Model{ID: 2}, Name: "Side Gate", Port: "COM4", BaudRate: 9600, Enabled: true})
	hardware.Manager.AddScale(models.ScaleConfig{Model: gorm.Model{ID: 3}, Name: "Back Gate", Port: "COM5", BaudRate: 9600, Enabled: true})

	// 3. Initialize CV
	anpr := cv.NewANPRService("models/platdetection.pt")

	// 4. Initialize Handlers
	server := handlers.NewServer(db, hardware.Manager, anpr)

	// 5. Setup Router
	r := gin.Default()

	// Static Files
	r.Static("/static", "./web/static")

	// HTML Templates
	// Gin requires loading templates before defining routes that use them
	r.LoadHTMLGlob("web/templates/*")

	// Routes
	r.GET("/", server.ShowDashboard)
	r.GET("/dashboard", server.ShowDashboard)
	r.GET("/weighing", server.ShowWeighing)

	api := r.Group("/api")
	{
		api.POST("/transaction", server.SaveTransaction)
		api.POST("/anpr/trigger", server.TriggerANPR)
		api.GET("/scales/stream", server.StreamScaleData)
	}

	log.Println("StoneWeigh Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Server failed:", err)
	}
}
