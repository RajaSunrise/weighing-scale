package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Password string `json:"-"` // Store hashed password
	Name     string `json:"name"`
	Role     string `json:"role"` // admin, operator
}

type Transaction struct {
	gorm.Model
	TicketID      string    `gorm:"uniqueIndex" json:"ticket_id"`
	PlateNumber   string    `json:"plate_number"`
	DriverName    string    `json:"driver_name"`
	Vendor        string    `json:"vendor"`
	Material      string    `json:"material"`
	InboundWeight float64   `json:"inbound_weight"` // Berat Masuk
	OutboundWeight float64  `json:"outbound_weight"` // Berat Keluar
	NetWeight     float64   `json:"net_weight"`
	SnapshotPath  string    `json:"snapshot_path"` // Path to image file
	Status        string    `json:"status"` // PENDING, COMPLETED
	EntryTime     time.Time `json:"entry_time"`
	ExitTime      time.Time `json:"exit_time"`
}
