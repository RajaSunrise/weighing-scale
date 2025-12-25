//go:build !gocv

package cv

import (
	"log"
	"strings"
)

// ANPRService handles license plate detection (Mock Version)
type ANPRService struct {
	IsLoaded  bool
	ModelPath string
}

func NewANPRService(modelPath string) *ANPRService {
	log.Printf("ANPR Service running in MOCK mode (No OpenCV detected)")
	return &ANPRService{
		IsLoaded:  true,
		ModelPath: modelPath,
	}
}

// CaptureAndDetect connects to a CCTV stream (RTSP) or camera ID and returns the detected text and snapshot path
func (s *ANPRService) CaptureAndDetect(cameraSource string) (string, string, error) {
	log.Printf("Mock ANPR: Capturing from %s", cameraSource)

	// Simulate correct results for known test images
	if strings.Contains(cameraSource, "test_image_1") {
		return "B 8187", "/static/images/placeholder_truck.jpg", nil
	}
	if strings.Contains(cameraSource, "test_image_2") {
		return "B 9190 IC", "/static/images/placeholder_truck.jpg", nil
	}
	if strings.Contains(cameraSource, "test_image_3") {
		return "K 8324 QD", "/static/images/placeholder_truck.jpg", nil
	}

	// Default simulation
	return "B 1234 MOCK", "/static/images/placeholder_truck.jpg", nil
}
