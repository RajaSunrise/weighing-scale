//go:build !gocv

package cv

import (
	"log"
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

	// Simulate success
	return "B 1234 MOCK", "/static/images/placeholder_truck.jpg", nil
}
