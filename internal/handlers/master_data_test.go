package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"stoneweigh/internal/models"
)

func TestCreateVehicle(t *testing.T) {
	r, db := setupServer(t)

	// Test 1: Create Vehicle (Success)
	payload := map[string]any{
		"plate_number":  "B 5555 NEW",
		"driver_name":   "New Driver",
		"default_tare":  8000.0,
		"owner_company": "Legacy Corp",
	}
	body, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/vehicles", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify DB
	var v models.Vehicle
	db.First(&v, "plate_number = ?", "B 5555 NEW")
	assert.Equal(t, "New Driver", v.DriverName)

	// Test 2: Duplicate
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/vehicles", bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusInternalServerError, w2.Code)
}

func TestUpdateVehicle(t *testing.T) {
	r, db := setupServer(t)

	// Seed
	v := models.Vehicle{PlateNumber: "B 1111", DriverName: "Old", DefaultTare: 1000}
	db.Create(&v)

	payload := map[string]any{
		"plate_number": "B 1111",
		"driver_name":  "Updated",
		"default_tare": 2000.0,
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/vehicles/%d", v.ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.Vehicle
	db.First(&updated, v.ID)
	assert.Equal(t, "Updated", updated.DriverName)
	assert.Equal(t, 2000.0, updated.DefaultTare)
}

func TestDeleteVehicle(t *testing.T) {
	r, db := setupServer(t)
	v := models.Vehicle{PlateNumber: "D 3333", DriverName: "Delete Me"}
	db.Create(&v)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/vehicles/%d", v.ID), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var count int64
	db.Model(&models.Vehicle{}).Where("id = ?", v.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestGetVehicleDetails(t *testing.T) {
	r, db := setupServer(t)
	v := models.Vehicle{PlateNumber: "B 1234 XX", DriverName: "Found Me"}
	db.Create(&v)

	// Success
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/vehicles/details?plate=B 1234 XX", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Found Me")

	// Not Found
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/vehicles/details?plate=UNKNOWN", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNotFound, w2.Code)
}

func TestSearchVehicles(t *testing.T) {
	r, db := setupServer(t)

	db.Create(&models.Vehicle{PlateNumber: "B 1000 X"})
	db.Create(&models.Vehicle{PlateNumber: "B 2000 X"})
	db.Create(&models.Vehicle{PlateNumber: "D 3000 Z"})

	// Search "B 1"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/vehicles/search?q=B 1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "B 1000 X")
	assert.NotContains(t, w.Body.String(), "B 2000 X") // Should not match

	// Empty Search
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/vehicles/search?q=", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "[]", w2.Body.String())
}

func TestCreateCompany(t *testing.T) {
	r, _ := setupServer(t)

	payload := map[string]string{
		"name":    "PT Stone",
		"address": "Mountain",
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/companies", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}
