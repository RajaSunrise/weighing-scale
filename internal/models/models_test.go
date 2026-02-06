package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(
		&WeighingRecord{},
		&WeighingStation{},
		&StationCamera{},
		&User{},
		&Vehicle{},
		&Company{},
		&UserStationAssignment{},
		&Invoice{},
	)
	assert.NoError(t, err)
	return db
}

func TestWeighingRecordValidation(t *testing.T) {
	db := setupTestDB(t)

	// Test Empty Plate
	r1 := WeighingRecord{DriverName: "Budi", GrossWeight: 100}
	err := db.Create(&r1).Error
	assert.Error(t, err)
	assert.Equal(t, "plate number tidak boleh kosong", err.Error())

	// Test Empty Driver
	r2 := WeighingRecord{PlateNumber: "B 123", GrossWeight: 100}
	err = db.Create(&r2).Error
	assert.Error(t, err)
	assert.Equal(t, "nama supir tidak boleh kosong", err.Error())

	// Test Zero Gross Weight
	r3 := WeighingRecord{PlateNumber: "B 123", DriverName: "Budi", GrossWeight: 0}
	err = db.Create(&r3).Error
	assert.Error(t, err)
	assert.Equal(t, "berat kotor tidak boleh kosong", err.Error())

	// Test Valid
	r4 := WeighingRecord{
		PlateNumber: "B 123",
		DriverName: "Budi",
		GrossWeight: 100,
		TicketNumber: "T-123",
	}
	err = db.Create(&r4).Error
	assert.NoError(t, err)
}

func TestWeighingRecordIndex(t *testing.T) {
	db := setupTestDB(t)
	migrator := db.Migrator()

	// Check indexes
	assert.True(t, migrator.HasIndex(&WeighingRecord{}, "WeighedAt"), "Index on WeighedAt should exist")
	assert.True(t, migrator.HasIndex(&WeighingRecord{}, "TicketNumber"), "Index on TicketNumber should exist")
	assert.True(t, migrator.HasIndex(&WeighingRecord{}, "PlateNumber"), "Index on PlateNumber should exist")
}

func TestStationCameraRelation(t *testing.T) {
	db := setupTestDB(t)

	station := WeighingStation{Name: "Main Gate", Enabled: true}
	db.Create(&station)

	cam1 := StationCamera{Name: "Cam 1", RTSPURL: "rtsp://192.168.1.1", WeighingStationID: station.ID}
	cam2 := StationCamera{Name: "Cam 2", RTSPURL: "rtsp://192.168.1.2", WeighingStationID: station.ID}
	db.Create(&cam1)
	db.Create(&cam2)

	// Fetch Station with Cameras
	var fetched WeighingStation
	err := db.Preload("Cameras").First(&fetched, station.ID).Error
	assert.NoError(t, err)
	assert.Len(t, fetched.Cameras, 2)
	assert.Equal(t, "Cam 1", fetched.Cameras[0].Name)
}

func TestUserConstraints(t *testing.T) {
	db := setupTestDB(t)

	u1 := User{Username: "admin", Role: "admin"}
	err := db.Create(&u1).Error
	assert.NoError(t, err)

	// Test Unique Username
	u2 := User{Username: "admin", Role: "operator"}
	err = db.Create(&u2).Error
	assert.Error(t, err) // Should fail unique constraint
}

func TestVehicleConstraints(t *testing.T) {
	db := setupTestDB(t)

	v1 := Vehicle{PlateNumber: "B 1234 XX", DefaultTare: 5000}
	err := db.Create(&v1).Error
	assert.NoError(t, err)

	// Test Unique Plate
	v2 := Vehicle{PlateNumber: "B 1234 XX", DefaultTare: 6000}
	err = db.Create(&v2).Error
	assert.Error(t, err)
}

func TestRTSPURLValidation(t *testing.T) {
	db := setupTestDB(t)

	// Test Valid URLs
	validURLs := []string{
		"rtsp://192.168.1.100:554/stream",
		"http://192.168.1.100/video.mjpg",
	}

	for _, u := range validURLs {
		cam := StationCamera{Name: "Cam Valid", RTSPURL: u}
		err := db.Create(&cam).Error
		assert.NoError(t, err, "Should accept valid URL: "+u)
	}

	// Test Invalid URLs
	invalidURLs := []string{
		"file:///etc/passwd",
		"ftp://example.com/file",
		"gopher://example.com",
		"javascript:alert(1)",
		"tcp://127.0.0.1:1234",    // Loopback blocked
		"rtsp://localhost:554",    // Localhost blocked
		"http://169.254.169.254/", // Link-local blocked
	}

	for _, u := range invalidURLs {
		cam := StationCamera{Name: "Cam Invalid", RTSPURL: u}
		err := db.Create(&cam).Error
		assert.Error(t, err, "Should reject invalid URL: "+u)
	}

	// Test Malformed URL
	cam := StationCamera{Name: "Cam Malformed", RTSPURL: "::not-a-url"}
	err := db.Create(&cam).Error
	assert.Error(t, err)

	// Test Legacy WeighingStation CameraURL
	wsValid := WeighingStation{Name: "WS Valid", CameraURL: "rtsp://192.168.1.50"}
	err = db.Create(&wsValid).Error
	assert.NoError(t, err)

	wsInvalid := WeighingStation{Name: "WS Invalid", CameraURL: "file:///invalid"}
	err = db.Create(&wsInvalid).Error
	assert.Error(t, err)
}
