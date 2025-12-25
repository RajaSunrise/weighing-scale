package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"stoneweigh/internal/database"
	"stoneweigh/internal/handlers"
)

func main() {
	// Initialize Database
	// Set default env vars for sandbox if not set
	if os.Getenv("DB_HOST") == "" {
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_USER", "postgres")
		os.Setenv("DB_PASSWORD", "postgres")
		os.Setenv("DB_NAME", "stoneweigh")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("USE_SQLITE", "true") // Fallback
	}

	database.Connect()

	r := gin.Default()

	// Load Templates
	// We need to define the layout logic.
	// Gin's HTML render defaults to loading all.
	// We'll use LoadHTMLGlob.
	r.LoadHTMLGlob("web/templates/*")

	// Serve Static Files
	r.Static("/static", "./web/static")

	// Routes
	r.GET("/", handlers.ShowLogin)
	r.GET("/dashboard", handlers.ShowDashboard)
	r.GET("/weighing-station", handlers.ShowWeighing)
	r.GET("/report-dashboard", handlers.ShowDashboard) // Placeholder re-use
	r.GET("/driver-vehicle", handlers.ShowDashboard) // Placeholder re-use
	r.GET("/user-management", handlers.ShowDashboard) // Placeholder re-use
	r.GET("/settings-hardware", handlers.ShowDashboard) // Placeholder re-use

	// API Routes
	api := r.Group("/api")
	{
		api.POST("/login", handlers.Login)
		api.GET("/transactions", handlers.GetTransactions)
		api.POST("/transactions", handlers.CreateTransaction)
		api.GET("/stream", handlers.StreamCCTV)
	}

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
