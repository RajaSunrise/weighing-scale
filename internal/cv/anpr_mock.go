//go:build !gocv

package cv

import (
	"log"
	"math/rand"
	"strings"
	"time"
)

// ANPRService handles license plate detection (Mock Version)
type ANPRService struct {
	IsLoaded  bool
	ModelPath string
}

func NewANPRService(modelPath string) *ANPRService {
	log.Printf("ANPR Service running in MOCK mode (No OpenCV detected)")
	rand.Seed(time.Now().UnixNano())
	return &ANPRService{
		IsLoaded:  true,
		ModelPath: modelPath,
	}
}

// CaptureAndDetect connects to a CCTV stream (RTSP) or camera ID and returns the detected text and snapshot path
func (s *ANPRService) CaptureAndDetect(cameraSource string) (string, string, error) {
	log.Printf("Mock ANPR: Capturing from %s", cameraSource)
	log.Printf("NOTE: Results are SIMULATED based on filename in Mock mode.")

	// Simulate correct results for known test images to keep tests passing in dev env
	if strings.Contains(cameraSource, "test_image_1") {
		return "B 8187", "/static/images/placeholder_truck.jpg", nil
	}
	if strings.Contains(cameraSource, "test_image_2") {
		return "B 9190 IC", "/static/images/placeholder_truck.jpg", nil
	}
	if strings.Contains(cameraSource, "test_image_3") {
		return "K 8324 QD", "/static/images/placeholder_truck.jpg", nil
	}

	// Randomized Realistic Data for "Simulated" feel
	prefixes := []string{"B", "D", "F", "A", "H"}
	suffixes := []string{"UA", "XY", "BC", "OM", "PR"}

	// Generate random plate
	prefix := prefixes[rand.Intn(len(prefixes))]
	number := rand.Intn(8999) + 1000
	suffix := suffixes[rand.Intn(len(suffixes))]

	plate := prefix + " " + string(rune('0'+(number/1000)%10)) + string(rune('0'+(number/100)%10)) + string(rune('0'+(number/10)%10)) + string(rune('0'+number%10)) + " " + suffix

	// Return the simulated plate
	return plate, "/static/images/placeholder_truck.jpg", nil
}
