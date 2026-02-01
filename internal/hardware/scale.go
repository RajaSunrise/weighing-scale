package hardware

import (
	"bufio"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
	Config         models.WeighingStation
	Port           serial.Port
	lastWeightBits uint64 // atomic float64
	connectedInt   int32  // atomic bool (0 or 1)
}

func (sc *ScaleConnection) SetWeight(w float64) {
	atomic.StoreUint64(&sc.lastWeightBits, math.Float64bits(w))
}

func (sc *ScaleConnection) GetWeight() float64 {
	return math.Float64frombits(atomic.LoadUint64(&sc.lastWeightBits))
}

func (sc *ScaleConnection) SetConnected(connected bool) {
	var val int32
	if connected {
		val = 1
	}
	atomic.StoreInt32(&sc.connectedInt, val)
}

func (sc *ScaleConnection) IsConnected() bool {
	return atomic.LoadInt32(&sc.connectedInt) == 1
}

var Manager *ScaleManager

func InitScaleManager() {
	Manager = &ScaleManager{
		Scales:    make(map[uint]*ScaleConnection),
		stopChans: make(map[uint]chan bool),
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

	// 2. Add or Update stations
	// For simplicity, we'll stop and restart even if unchanged,
	// or we could check diffs. Let's restart to ensure clean state.
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
		conn.SetWeight(weight)
		conn.SetConnected(connected)
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

		if !conn.IsConnected() {
			// Attempt connection
			// Default serial settings if not specified
			baud := conn.Config.BaudRate
			if baud == 0 {
				baud = 9600
			}

			mode := &serial.Mode{
				BaudRate: baud,
				DataBits: 8,
				Parity:   serial.NoParity,
				StopBits: serial.OneStopBit,
			}

			port, err := serial.Open(conn.Config.ScalePort, mode)
			if err != nil {
				// Failed to connect, wait and retry
				conn.SetConnected(false)
				conn.SetWeight(0)

				// Sleep with check for stop
				select {
				case <-time.After(5 * time.Second):
				case <-stopChan:
					return
				}
				continue
			}

			conn.Port = port
			conn.SetConnected(true)
			log.Printf("Connected to Scale %d (%s) on %s", scaleID, conn.Config.Name, conn.Config.ScalePort)
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

			// OPTIMIZATION: Update directly using atomics to avoid Global Lock contention
			conn.SetWeight(weight)
			conn.SetConnected(true)
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading scale %d: %v", scaleID, err)
			conn.Port.Close()
			conn.SetConnected(false)
			conn.SetWeight(0)
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
				if !conn.IsConnected() {
					// Toggle between empty (0) and loaded (~25000)
					now := time.Now().Unix()
					if (now/20)%2 == 0 {
						conn.SetWeight(0)
					} else {
						// Jitter
						conn.SetWeight(24500 + float64(now%100))
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
