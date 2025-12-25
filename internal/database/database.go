package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	"stoneweigh/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	var err error

	// Check if user explicitly wants SQLite or if defaults to Postgres
	// In the original code, it defaulted to Postgres unless USE_SQLITE was true.
	// But the user complained about silent fallback.
	// We will change the logic:
	// 1. If DB_DRIVER is "sqlite", use SQLite.
	// 2. If DB_DRIVER is "postgres" (or default/missing), use Postgres.
	// 3. IF Postgres fails, DO NOT FALLBACK. Return/Log Fatal error.

	driver := strings.ToLower(os.Getenv("DB_DRIVER"))
	if driver == "sqlite" {
		log.Println("Using SQLite database (configured)...")
		DB, err = gorm.Open(sqlite.Open("stoneweigh.db"), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to SQLite: %v", err)
		}
	} else {
		// Default to Postgres
		// Parse DSN or build from individual vars. The .env.example shows DB_DSN usage for Postgres too,
		// but the code used individual vars. Let's support both or stick to the code's pattern.
		// The original code used: host=%s user=%s ...
		// Let's stick to the code's pattern but also allow DB_DSN overriding.

		var dsn string
		if os.Getenv("DB_DSN") != "" && !strings.HasSuffix(os.Getenv("DB_DSN"), ".db") {
			dsn = os.Getenv("DB_DSN")
		} else {
			dsn = fmt.Sprintf(
				"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
				os.Getenv("DB_HOST"),
				os.Getenv("DB_USER"),
				os.Getenv("DB_PASSWORD"),
				os.Getenv("DB_NAME"),
				os.Getenv("DB_PORT"),
			)
		}

		log.Println("Attempting to connect to Postgres...")
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			// CRITICAL CHANGE: No silent fallback.
			log.Fatalf("Failed to connect to Postgres: %v. Please check your configuration.", err)
		}
	}

	log.Println("Database connected successfully")

	// Migration
	DB.AutoMigrate(
		&models.User{},
		&models.Invoice{},
		&models.ScaleConfig{},
		&models.Vehicle{},
		&models.WeighingRecord{},
		&models.WeighingStation{},
		&models.UserStationAssignment{},
	)
}
