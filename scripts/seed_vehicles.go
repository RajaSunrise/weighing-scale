package main

import (
	"log"
	"stoneweigh/internal/database"
	"stoneweigh/internal/models"
)

func main() {
	database.Connect()
	db := database.DB

	vehicles := []models.Vehicle{
		{PlateNumber: "B 1234 XX", DriverName: "Test Driver"},
		{PlateNumber: "B 5678 YY", DriverName: "Another Driver"},
        {PlateNumber: "D 9999 ZZ", DriverName: "Bandung Driver"},
	}

	for _, v := range vehicles {
		if err := db.Create(&v).Error; err != nil {
			log.Printf("Skipping %s: %v", v.PlateNumber, err)
		} else {
			log.Printf("Seeded %s", v.PlateNumber)
		}
	}
}
