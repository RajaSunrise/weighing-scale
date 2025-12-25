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

// ScaleConfig represents the configuration for a physical scale connected via Serial
type ScaleConfig struct {
	gorm.Model
	Name       string `json:"name"`        // e.g., "Main Gate Scale"
	Port       string `json:"port"`        // e.g., "COM3" or "/dev/ttyUSB0"
	BaudRate   int    `json:"baud_rate"`   // e.g., 9600
	DataBits   int    `json:"data_bits"`   // e.g., 8
	StopBits   int    `json:"stop_bits"`   // e.g., 1
	Parity     int    `json:"parity"`      // 0:None, 1:Odd, 2:Even
	Enabled    bool   `json:"enabled"`
}

// Vehicle represents master data for known vehicles (optional but useful for frequent trucks)
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
