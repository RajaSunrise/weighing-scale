package models

import (
	"time"

	"gorm.io/gorm"
)

// WeighingRecord represents a single weighing transaction
type WeighingRecord struct {
	gorm.Model
	TicketNumber string    `gorm:"uniqueIndex;not null" json:"ticket_number"`
	ScaleID      uint      `json:"scale_id"`
	PlateNumber  string    `gorm:"index" json:"plate_number"`
	DriverName   string    `json:"driver_name"`
	ManagerName  string    `json:"manager_name"` // Name of the operator/manager
	Product      string    `json:"product"`

	GrossWeight  float64   `json:"gross_weight"` // Initial weight
	TareWeight   float64   `json:"tare_weight"`  // Empty weight
	NetWeight    float64   `json:"net_weight"`   // Gross - Tare

	Status       string    `json:"status"` // "PENDING", "COMPLETED", "VOID"

	// Snapshots paths
	SnapshotFront string `json:"snapshot_front"` // CCTV Path
	SnapshotBack  string `json:"snapshot_back"`  // CCTV Path
	InvoicePath   string `json:"invoice_path"`   // PDF Path

	WeighedAt     time.Time `json:"weighed_at"`
}

// WeighingStation represents a physical weighing station configuration
// It combines Scale config and Camera config into one logical unit.
type WeighingStation struct {
	gorm.Model
	Name       string `json:"name"`        // e.g., "Main Gate"
	ScalePort  string `json:"scale_port"`  // e.g., "COM3" or "/dev/ttyUSB0"
	BaudRate   int    `json:"baud_rate"`   // e.g., 9600
	CameraURL  string `json:"camera_url"`  // RTSP URL
	Enabled    bool   `json:"enabled"`
}

// Deprecated: Use WeighingStation instead. Kept for migration safety if needed,
// but we will likely migrate data to WeighingStation.
type ScaleConfig struct {
	gorm.Model
	Name       string `json:"name"`
	Port       string `json:"port"`
	BaudRate   int    `json:"baud_rate"`
	DataBits   int    `json:"data_bits"`
	StopBits   int    `json:"stop_bits"`
	Parity     int    `json:"parity"`
	Enabled    bool   `json:"enabled"`
}

// Vehicle represents master data for known vehicles
type Vehicle struct {
	gorm.Model
	PlateNumber   string  `gorm:"uniqueIndex" json:"plate_number"`
	DriverName    string  `json:"driver_name"`
	DefaultTare   float64 `json:"default_tare"` // Known empty weight
	OwnerCompany  string  `json:"owner_company"`
}

// Invoice metadata
type Invoice struct {
	gorm.Model
	WeighingRecordID uint           `json:"weighing_record_id"`
	WeighingRecord   WeighingRecord `json:"weighing_record"`
	InvoiceNumber    string         `gorm:"uniqueIndex" json:"invoice_number"`
	Amount           float64        `json:"amount"` // Calculated cost
	GeneratedAt      time.Time      `json:"generated_at"`
}

// User represents a system user (Admin/Operator)
type User struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string `json:"-"` // Store bcrypt hash
	FullName     string `json:"full_name"`
	Role         string `json:"role"` // "admin", "operator"
}

// UserStationAssignment links a User to specific WeighingStations.
// If a user has NO assignments, they might see nothing (or all, depending on policy).
// We will enforce: No assignment = No access to operate.
type UserStationAssignment struct {
	gorm.Model
	UserID            uint            `json:"user_id"`
	User              User            `json:"user"`
	WeighingStationID uint            `json:"weighing_station_id"`
	WeighingStation   WeighingStation `json:"weighing_station"`
}
