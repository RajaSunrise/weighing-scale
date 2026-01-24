package hardware

import (
	"bufio"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
	"gorm.io/gorm"
	"stoneweigh/internal/models"
)

// ScaleManager handles connections to multiple scales
type ScaleManager struct {
	Scales      map[uint]*ScaleConnection
	Mu          sync.Mutex
	stopChans   map[uint]chan bool // To stop monitoring goroutines
}

type ScaleConnection struct {
	Config models.WeighingStation
	Port   serial.Port
	LastWeight float64
	Connected  bool
}

var Manager *ScaleManager

func InitScaleManager() {
	Manager = &ScaleManager{
		Scales:      make(map[uint]*ScaleConnection),
		stopChans:   make(map[uint]chan bool),
	}
}

// ReloadConfig loads configuration from the DB and restarts connections
func (sm *ScaleManager) ReloadConfig(db *gorm.DB) {
	log.Println("Reloading Scale Configurations...")
	var stations []models.WeighingStation
	if err := db.Where("enabled = ?", true).Find(&stations).Error; err != nil {
		log.Printf("Error loading stations: %v", err)
		return
	}

	// 1. Identify removed or updated stations
	sm.Mu.Lock()
	currentIDs := make(map[uint]bool)
	for _, s := range stations {
		currentIDs[s.ID] = true
	}

	// Stop monitors for stations that no longer exist or are disabled
	for id, _ := range sm.Scales {
		if !currentIDs[id] {
			if stop, ok := sm.stopChans[id]; ok {
				close(stop)
				delete(sm.stopChans, id)
			}
			if conn, ok := sm.Scales[id]; ok && conn.Port != nil {
				conn.Port.Close()
			}
			delete(sm.Scales, id)
			log.Printf("Stopped Scale %d", id)
		}
	}
	sm.Mu.Unlock()

	for _, station := range stations {
		sm.AddOrUpdateScale(station)
	}
}

// AddOrUpdateScale registers and attempts to connect to a scale
func (sm *ScaleManager) AddOrUpdateScale(config models.WeighingStation) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	// If exists, stop first
	if _, exists := sm.Scales[config.ID]; exists {
		if stop, ok := sm.stopChans[config.ID]; ok {
			close(stop)
			delete(sm.stopChans, config.ID)
		}
		// Close port if open
		if sm.Scales[config.ID].Port != nil {
			sm.Scales[config.ID].Port.Close()
		}
	}

	conn := &ScaleConnection{
		Config: config,
	}
	sm.Scales[config.ID] = conn

	stop := make(chan bool)
	sm.stopChans[config.ID] = stop

	go sm.monitorScale(config.ID, stop)
}

// UpdateScale safely updates the scale state
func (sm *ScaleManager) UpdateScale(id uint, weight float64, connected bool) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()
	if conn, ok := sm.Scales[id]; ok {
		conn.LastWeight = weight
		conn.Connected = connected
	}
}

// monitorScale constantly tries to read from the scale
func (sm *ScaleManager) monitorScale(scaleID uint, stopChan chan bool) {
	for {
		select {
		case <-stopChan:
			return
		default:
			// Continue
		}

		sm.Mu.Lock()
		conn, exists := sm.Scales[scaleID]
		sm.Mu.Unlock()

		if !exists {
			return
		}

		// Check if PORT is nil, regardless of "Connected" status (which might be set by remote API)
		if conn.Port == nil {
			// Attempt connection
			// Default serial settings if not specified
			baud := conn.Config.BaudRate
			if baud == 0 { baud = 9600 }

			mode := &serial.Mode{
				BaudRate: baud,
				DataBits: 8,
				Parity:   serial.NoParity,
				StopBits: serial.OneStopBit,
			}

			port, err := serial.Open(conn.Config.ScalePort, mode)
			if err != nil {

				// Sleep with check for stop
				select {
				case <-time.After(5 * time.Second):
				case <-stopChan:
					return
				}
				continue
			}

			conn.Port = port

			sm.UpdateScale(scaleID, conn.LastWeight, true) // Keep last weight? Or 0?
			log.Printf("Connected to Scale %d (%s) on %s", scaleID, conn.Config.Name, conn.Config.ScalePort)
		}

		// Double check Port is not nil before creating scanner
		if conn.Port == nil {
			continue
		}

		// Read loop
		scanner := bufio.NewScanner(conn.Port)
		for scanner.Scan() {
			select {
			case <-stopChan:
				conn.Port.Close()
				return
			default:
			}

			text := scanner.Text()
			weight := parseWeight(text)

			sm.UpdateScale(scaleID, weight, true)
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading scale %d: %v", scaleID, err)
			conn.Port.Close()
			conn.Port = nil // Ensure we reset to nil so loop tries to reconnect
			sm.UpdateScale(scaleID, 0, false)
		} else {
			// EOF or explicit break (unlikely with serial unless closed)
			conn.Port.Close()
			conn.Port = nil
		}
	}
}

// Demo Mode: Simulates scale activity
func (sm *ScaleManager) StartDemoMode() {
	go func() {
		log.Println("Starting Demo Scale Simulation...")
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			sm.Mu.Lock()
			// Simulate random weights for Scale 1 (or any existing scale)
			// Iterate through all scales and simulate if not connected
			for _, conn := range sm.Scales {
				if !conn.Connected {
					// Toggle between empty (0) and loaded (~25000)
					now := time.Now().Unix()
					if (now/20)%2 == 0 {
						conn.LastWeight = 0
					} else {
						// Jitter
						conn.LastWeight = 24500 + float64(now%100)
					}
				}
			}
			sm.Mu.Unlock()
		}
	}()
}

func parseWeight(raw string) float64 {
	clean := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || r == '.' || r == '-' {
			return r
		}
		return -1
	}, raw)

	if val, err := strconv.ParseFloat(clean, 64); err == nil {
		return val
	}
	return 0.0
}
