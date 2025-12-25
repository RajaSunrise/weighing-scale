package hardware

import (
	"bufio"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
	"stoneweigh/internal/models"
)

// ScaleManager handles connections to multiple scales
type ScaleManager struct {
	Scales      map[uint]*ScaleConnection
	DataChannel chan ScaleData // Channel to broadcast updates
	Mu          sync.Mutex
}

type ScaleConnection struct {
	Config models.ScaleConfig
	Port   serial.Port
	LastWeight float64
	Connected  bool
}

type ScaleData struct {
	ScaleID   uint    `json:"scale_id"`
	Weight    float64 `json:"weight"`
	Connected bool    `json:"connected"`
	Timestamp int64   `json:"timestamp"`
}

var Manager *ScaleManager

func InitScaleManager() {
	Manager = &ScaleManager{
		Scales:      make(map[uint]*ScaleConnection),
		DataChannel: make(chan ScaleData, 100),
	}
}

// AddScale registers and attempts to connect to a scale
func (sm *ScaleManager) AddScale(config models.ScaleConfig) {
	sm.Mu.Lock()
	defer sm.Mu.Unlock()

	conn := &ScaleConnection{
		Config: config,
	}
	sm.Scales[config.ID] = conn
	go sm.monitorScale(config.ID)
}

// monitorScale constantly tries to read from the scale
func (sm *ScaleManager) monitorScale(scaleID uint) {
	for {
		sm.Mu.Lock()
		conn, exists := sm.Scales[scaleID]
		sm.Mu.Unlock()

		if !exists {
			return
		}

		if !conn.Connected {
			// Attempt connection
			mode := &serial.Mode{
				BaudRate: conn.Config.BaudRate,
				DataBits: conn.Config.DataBits,
				Parity:   serial.Parity(conn.Config.Parity),
				StopBits: serial.StopBits(conn.Config.StopBits),
			}

			port, err := serial.Open(conn.Config.Port, mode)
			if err != nil {
				// Failed to connect, wait and retry
				// log.Printf("Failed to open scale %d on %s: %v", scaleID, conn.Config.Port, err)

				// Send disconnected status
				sm.DataChannel <- ScaleData{ScaleID: scaleID, Connected: false, Timestamp: time.Now().Unix()}
				time.Sleep(5 * time.Second)
				continue
			}

			conn.Port = port
			conn.Connected = true
			log.Printf("Connected to Scale %d on %s", scaleID, conn.Config.Port)
		}

		// Read loop
		scanner := bufio.NewScanner(conn.Port)
		for scanner.Scan() {
			text := scanner.Text()
			weight := parseWeight(text)

			conn.LastWeight = weight

			// Broadcast
			sm.DataChannel <- ScaleData{
				ScaleID:   scaleID,
				Weight:    weight,
				Connected: true,
				Timestamp: time.Now().Unix(),
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading scale %d: %v", scaleID, err)
			conn.Port.Close()
			conn.Connected = false
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
			// Simulate random weights for Scale 1 if disconnected
			if conn, ok := sm.Scales[1]; ok {
				if !conn.Connected {
					// Toggle between empty (0) and loaded (~25000)
					now := time.Now().Unix()
					if (now/20)%2 == 0 {
						conn.LastWeight = 0
					} else {
						// Jitter
						conn.LastWeight = 24500 + float64(now%100)
					}
					// Do not set conn.Connected = true here to avoid confusing monitorScale
				}

				// Broadcast fake data
				sm.DataChannel <- ScaleData{
					ScaleID:   1,
					Weight:    conn.LastWeight,
					Connected: true, // Tell frontend it's connected
					Timestamp: time.Now().Unix(),
				}
			}
			sm.Mu.Unlock()
		}
	}()
}

// parseWeight parses the raw serial string from a generic scale indicator
// This varies wildly by manufacturer (Mettler Toledo, Avery, etc.)
// For this MVP, we assume a simple format or just extract numbers.
func parseWeight(raw string) float64 {
	// Example format: "ST,GS,  12040 kg" or "  12040"
	// Sanitize string to keep only numbers and dot
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
