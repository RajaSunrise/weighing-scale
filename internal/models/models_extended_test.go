package models

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Tests in the same package can share helper functions like setupTestDB

func TestCompanyValidation(t *testing.T) {
	db := setupTestDB(t)

	// Test 1: Valid Company
	c1 := Company{Name: "PT Maju Jaya", Address: "Jl. Sudirman", Phone: "08123456"}
	err := db.Create(&c1).Error
	assert.NoError(t, err)

	// Test 2: Missing Name (Should fail due to not null)
	c2 := Company{Address: "No Name St"}
	err = db.Create(&c2).Error
	if assert.Error(t, err) {
		// We now enforce this via BeforeCreate hook or DB constraint
		// "company name cannot be empty" or "NOT NULL constraint"
		assert.True(t, strings.Contains(err.Error(), "cannot be empty") || strings.Contains(err.Error(), "NOT NULL"), "Error was: %v", err)
	}

	// Test 3: Duplicate Name (Unique Index)
	c3 := Company{Name: "PT Maju Jaya", Address: "Different Addr"}
	err = db.Create(&c3).Error
	assert.Error(t, err)
	// SQLite error message varies, usually "UNIQUE constraint failed"
}

func TestWeighingRecordCalculations(t *testing.T) {
	db := setupTestDB(t)

	now := time.Now()
	wr := WeighingRecord{
		TicketNumber: "T-CALC-001",
		PlateNumber:  "B 9999 XX",
		DriverName:   "Calculator",
		GrossWeight:  25000,
		TareWeight:   10000,
		// NetWeight not set manually, but usually logic calculates it BEFORE saving in Handler.
		// However, we should verify the DB stores what we give it.
		// If logic was in BeforeSave, we'd test it here.
		// Current logic: Handlers calculate Net. Models just store.
		NetWeight: 15000,
		WeighedAt: now,
	}

	err := db.Create(&wr).Error
	assert.NoError(t, err)

	var fetched WeighingRecord
	err = db.First(&fetched, "ticket_number = ?", "T-CALC-001").Error
	assert.NoError(t, err)
	assert.Equal(t, 15000.0, fetched.NetWeight)
}

func TestUserStationAssignment(t *testing.T) {
	db := setupTestDB(t)

	// Setup: User and Station
	user := User{Username: "operator1", Role: "operator"}
	db.Create(&user)

	station := WeighingStation{Name: "Station A", Enabled: true}
	db.Create(&station)

	// Assignment
	assign := UserStationAssignment{
		UserID:            user.ID,
		WeighingStationID: station.ID,
	}
	err := db.Create(&assign).Error
	assert.NoError(t, err)

	// Verify Preloading
	var loadedAssign UserStationAssignment
	err = db.Preload("User").Preload("WeighingStation").First(&loadedAssign, assign.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "operator1", loadedAssign.User.Username)
	assert.Equal(t, "Station A", loadedAssign.WeighingStation.Name)
}

func TestStationCameraValidationDetailed(t *testing.T) {
	db := setupTestDB(t)

	station := WeighingStation{Name: "Station Cam Test"}
	db.Create(&station)

	// Test Valid Schemes
	schemes := []string{"rtsps", "http", "https", "tcp", "udp"}
	for _, s := range schemes {
		cam := StationCamera{
			Name:              "Cam " + s,
			RTSPURL:           s + "://1.2.3.4",
			WeighingStationID: station.ID,
		}
		err := db.Create(&cam).Error
		assert.NoError(t, err, "Scheme %s should be valid", s)
	}

	// Test Invalid Schemes
	invalids := []string{"ftp", "file", "ws", "wss"}
	for _, s := range invalids {
		cam := StationCamera{
			Name:              "Cam " + s,
			RTSPURL:           s + "://1.2.3.4",
			WeighingStationID: station.ID,
		}
		err := db.Create(&cam).Error
		assert.Error(t, err, "Scheme %s should be invalid", s)
	}
}
