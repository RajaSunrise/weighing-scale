package models_test

import (
	"testing"
	"time"

	"stoneweigh/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWeighingRecordIndex(t *testing.T) {
	// 1. Connect to in-memory SQLite
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 2. Migrate
	err = db.AutoMigrate(&models.WeighingRecord{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// 3. Verify Index Exists
	if !db.Migrator().HasIndex(&models.WeighingRecord{}, "WeighedAt") {
		t.Errorf("Expected index on WeighedAt not found")
	}

    // Also try inserting to ensure no constraints broken
    rec := models.WeighingRecord{
        TicketNumber: "T-001",
        PlateNumber: "B 1234 XX",
        DriverName: "Budi",
        GrossWeight: 1000,
        WeighedAt: time.Now(),
    }
    if err := db.Create(&rec).Error; err != nil {
         t.Fatalf("Failed to create record: %v", err)
    }
}
