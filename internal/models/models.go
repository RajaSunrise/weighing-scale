package models

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
)

// WeighingRecord represents a single weighing transaction
type WeighingRecord struct {
	gorm.Model
	TicketNumber string `gorm:"uniqueIndex;not null" json:"ticket_number"`
	ScaleID      uint   `json:"scale_id"`
	PlateNumber  string `gorm:"index;not null" json:"plate_number"`
	DriverName   string `gorm:"not null" json:"driver_name"`
	CompanyName  string `json:"company_name"` // Owner/Company
	ManagerName  string `json:"manager_name"` // Name of the operator/manager
	Product      string `json:"product"`

	GrossWeight float64 `gorm:"not null" json:"gross_weight"` // Initial weight
	TareWeight  float64 `json:"tare_weight"`                  // Empty weight
	NetWeight   float64 `json:"net_weight"`                   // Gross - Tare

	Status string `json:"status"` // "PENDING", "COMPLETED", "VOID"

	// Snapshots paths
	SnapshotFront string `json:"snapshot_front"` // CCTV Path
	SnapshotBack  string `json:"snapshot_back"`  // CCTV Path
	InvoicePath   string `json:"invoice_path"`   // PDF Path

	// Indexed for faster range queries in reporting
	WeighedAt time.Time `gorm:"index" json:"weighed_at"`
}

func (wr *WeighingRecord) BeforeCreate(tx *gorm.DB) error {
	if wr.PlateNumber == "" {
		return errors.New("plate number tidak boleh kosong")
	}
	if wr.DriverName == "" {
		return errors.New("nama supir tidak boleh kosong")
	}
	if wr.GrossWeight == 0 {
		return errors.New("berat kotor tidak boleh kosong")
	}
	return nil
}

// WeighingStation represents a physical weighing station configuration
// It combines Scale config and Camera config into one logical unit.
type StationCamera struct {
	gorm.Model
	WeighingStationID uint            `json:"weighing_station_id"`
	WeighingStation   WeighingStation `json:"-"` // Prevent circular JSON
	Name              string          `json:"name"`
	RTSPURL           string          `json:"rtsp_url"`
}

func (sc *StationCamera) BeforeSave(tx *gorm.DB) error {
	return validateRTSPURL(sc.RTSPURL)
}

type WeighingStation struct {
	gorm.Model
	Name      string          `json:"name"`       // e.g., "Main Gate"
	ScalePort string          `json:"scale_port"` // e.g., "COM3" or "/dev/ttyUSB0"
	BaudRate  int             `json:"baud_rate"`  // e.g., 9600
	Cameras   []StationCamera `json:"cameras"`    // Multiple CCTVs
	Enabled   bool            `json:"enabled"`
	Token     string          `json:"token"`      // Security token for remote data push

	// Deprecated: Kept for migration, assume data moved to Cameras[0]
	CameraURL string `json:"camera_url,omitempty"`
}

func (ws *WeighingStation) BeforeSave(tx *gorm.DB) error {
	if ws.CameraURL != "" {
		return validateRTSPURL(ws.CameraURL)
	}
	return nil
}

func validateRTSPURL(rawURL string) error {
	if rawURL == "" {
		return nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return errors.New("invalid URL format")
	}

	scheme := strings.ToLower(parsed.Scheme)
	allowed := map[string]bool{
		"rtsp":  true,
		"rtsps": true,
		"http":  true,
		"https": true,
		"tcp":   true,
		"udp":   true,
	}

	if !allowed[scheme] {
		return errors.New("invalid URL scheme: must be rtsp, rtsps, http, https, tcp, or udp")
	}
	return nil
}

type ScaleConfig struct {
	gorm.Model
	Name     string `json:"name"`
	Port     string `json:"port"`
	BaudRate int    `json:"baud_rate"`
	DataBits int    `json:"data_bits"`
	StopBits int    `json:"stop_bits"`
	Parity   int    `json:"parity"`
	Enabled  bool   `json:"enabled"`
}

// Company represents a transport company
type Company struct {
	gorm.Model
	Name          string `gorm:"uniqueIndex;not null" json:"name"`
	Address       string `json:"address"`
	ContactPerson string `json:"contact_person"`
	Phone         string `json:"phone"`
}

func (c *Company) BeforeCreate(tx *gorm.DB) error {
	if strings.TrimSpace(c.Name) == "" {
		return errors.New("company name cannot be empty")
	}
	return nil
}

// Vehicle represents master data for known vehicles
type Vehicle struct {
	gorm.Model
	PlateNumber  string  `gorm:"uniqueIndex" json:"plate_number"`
	DriverName   string  `json:"driver_name"`
	DefaultTare  float64 `json:"default_tare"` // Known empty weight
	OwnerCompany string  `json:"owner_company"`

	SIM          string  `json:"sim"`
	CompanyID    *uint   `json:"company_id"`
	Company      Company `json:"company"`
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
