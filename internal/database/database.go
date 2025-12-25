package database

import (
	"fmt"
	"log"
	"os"

	"stoneweigh/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	// For production, use Postgres
	// DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	// For this prototype sandbox (where we might not have a running Postgres instance),
	// we would ideally use SQLite if Postgres fails, but sticking to the user request.
	// I will add a check.

	if os.Getenv("USE_SQLITE") == "true" {
		// Fallback for local testing without Postgres
		// You would need to import sqlite driver
		log.Println("Using SQLite (not implemented in this snippet, defaulting to Postgres attempt)")
	}

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return
	}

	log.Println("Database connected")
	DB.AutoMigrate(&models.User{}, &models.Transaction{})
}
