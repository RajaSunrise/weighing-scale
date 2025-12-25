package database

import (
	"fmt"
	"log"
	"os"

	"stoneweigh/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	var err error

	// Check for SQLite preference or fallback
	if os.Getenv("USE_SQLITE") == "true" {
		log.Println("Using SQLite database...")
		DB, err = gorm.Open(sqlite.Open("stoneweigh.db"), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to SQLite: %v", err)
		}
	} else {
		// Default to Postgres
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)
		log.Println("Attempting to connect to Postgres...")
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Printf("Failed to connect to Postgres: %v. Falling back to SQLite.", err)
			DB, err = gorm.Open(sqlite.Open("stoneweigh.db"), &gorm.Config{})
			if err != nil {
				log.Fatalf("Failed to fallback to SQLite: %v", err)
			}
		}
	}

	log.Println("Database connected successfully")
	DB.AutoMigrate(&models.User{}, &models.Invoice{}, &models.ScaleConfig{}, &models.Vehicle{}, &models.WeighingRecord{})
}
