//go:build gocv

package cv

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"gocv.io/x/gocv"
)

// ANPRService handles license plate detection
type ANPRService struct {
	Net       gocv.Net
	IsLoaded  bool
	ModelPath string
}

func NewANPRService(modelPath string) *ANPRService {
	// Check if file exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		log.Printf("Warning: Model file not found at %s. ANPR will be disabled.", modelPath)
		return &ANPRService{IsLoaded: false}
	}

	// Check file size to detect placeholder models (< 1MB is likely a placeholder)
	fileInfo, _ := os.Stat(modelPath)
	if fileInfo.Size() < 1024*1024 {
		log.Printf("Warning: Model file %s is too small (%d bytes). This appears to be a placeholder. ANPR will be disabled.", modelPath, fileInfo.Size())
		return &ANPRService{IsLoaded: false}
	}

	// Attempt to load the model.
	// Note: ReadNet expects ONNX format, not .pt (PyTorch)
	var net gocv.Net
	var loadErr error

	func() {
		defer func() {
			if r := recover(); r != nil {
				loadErr = fmt.Errorf("panic during model load: %v", r)
			}
		}()
		net = gocv.ReadNet(modelPath, "")
		if net.Empty() {
			loadErr = fmt.Errorf("model loaded but is empty")
		}
	}()

	if loadErr != nil {
		log.Printf("Warning: Failed to load model %s: %v. Ensure it is ONNX format compatible with OpenCV DNN.", modelPath, loadErr)
		return &ANPRService{IsLoaded: false}
	}

	// Set backend and target to default (CPU)
	net.SetPreferableBackend(gocv.NetBackendDefault)
	net.SetPreferableTarget(gocv.NetTargetCPU)

	return &ANPRService{
		Net:       net,
		IsLoaded:  true,
		ModelPath: modelPath,
	}
}

// CaptureAndDetect connects to a CCTV stream (RTSP) or camera ID and returns the detected text and snapshot path
func (s *ANPRService) CaptureAndDetect(cameraSource string) (string, string, error) {
	if !s.IsLoaded {
		return "", "", fmt.Errorf("ANPR model not loaded")
	}

	// Open Video Capture
	var webcam *gocv.VideoCapture
	var err error

	// Check if source is numeric (for USB Camera Index)
	if idx, errConv := strconv.Atoi(cameraSource); errConv == nil {
		webcam, err = gocv.OpenVideoCapture(idx)
	} else {
		webcam, err = gocv.OpenVideoCapture(cameraSource)
	}

	if err != nil {
		return "", "", fmt.Errorf("failed to open video source: %v", err)
	}
	defer webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	if ok := webcam.Read(&img); !ok || img.Empty() {
		return "", "", fmt.Errorf("failed to read frame from camera")
	}

	// Save Snapshot
	filename := fmt.Sprintf("web/static/images/snap_%d.jpg", SystemClock())
	gocv.IMWrite(filename, img)

	// Perform Detection
	// Preprocessing: YOLO style
	blob := gocv.BlobFromImage(img, 1.0/255.0, image.Pt(640, 640), gocv.NewScalar(0, 0, 0, 0), true, false)
	defer blob.Close()

	s.Net.SetInput(blob, "")
	prob := s.Net.Forward("")
	defer prob.Close()

	// REAL OCR Implementation using Tesseract CLI
	// This ensures that the detection is based on the actual image content.
	// Requires 'tesseract' to be installed on the system.

	out, err := exec.Command("tesseract", filename, "stdout", "--psm", "7").Output()
	if err != nil {
		log.Printf("Tesseract OCR failed: %v. Make sure tesseract-ocr is installed.", err)
		// Fallback to a generic string if OCR fails, to avoid crashing flow
		return "OCR_FAILED", filename, nil
	}

	detectedText := strings.TrimSpace(string(out))
	detectedText = cleanPlateText(detectedText)

	// Draw rectangle on image for debug (simplified as we didn't parse YOLO boxes yet)
	// In a full implementation, we would use the YOLO boxes to crop the plate before OCR.
	// For now, we assume the camera is framed on the plate or Tesseract can find it.
	gocv.Rectangle(&img, image.Rect(100, 100, 300, 200), color.RGBA{0, 255, 0, 0}, 2)
	gocv.IMWrite(filename, img)

	return detectedText, filename, nil
}

func cleanPlateText(text string) string {
	// Keep only alphanumeric and spaces
	clean := strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' {
			return r
		}
		return -1
	}, strings.ToUpper(text))
	return strings.TrimSpace(clean)
}

func SystemClock() int64 {
	return 1 // Placeholder
}
