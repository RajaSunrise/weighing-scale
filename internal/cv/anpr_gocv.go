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
	"time"

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
	// Note: ReadNet expects ONNX format, not .pt (PyTorch).
	// This service now assumes the model is a YOLOv8 ONNX model with
	// output shape roughly [1, 5, 8400] (for 1 class).
	// Users must convert their .pt model to ONNX using `yolo export model=platdetection.pt format=onnx`.
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

	// Perform Detection
	// Preprocessing: YOLO style (640x640)
	blob := gocv.BlobFromImage(img, 1.0/255.0, image.Pt(640, 640), gocv.NewScalar(0, 0, 0, 0), true, false)
	defer blob.Close()

	s.Net.SetInput(blob, "")
	prob := s.Net.Forward("")
	defer prob.Close()

	// Parse YOLOv8 Output
	// Output shape is typically [1, 5, 8400] for 1 class (x, y, w, h, score)
	// or [1, 4+nc, 8400]
	// We need to parse this to find the best bounding box.

	bestBox, found := processYOLOv8Output(prob, img.Cols(), img.Rows())

	var filename string
	var detectedText string

	if found {
		// Crop the license plate
		// Ensure coordinates are within bounds
		rect := bestBox
		if rect.Min.X < 0 { rect.Min.X = 0 }
		if rect.Min.Y < 0 { rect.Min.Y = 0 }
		if rect.Max.X > img.Cols() { rect.Max.X = img.Cols() }
		if rect.Max.Y > img.Rows() { rect.Max.Y = img.Rows() }

		// Draw rectangle for debug on original
		gocv.Rectangle(&img, rect, color.RGBA{0, 255, 0, 0}, 2)

		// Create cropped region for OCR
		region := img.Region(rect)
		defer region.Close()

		// Resize for better OCR (upscale)
		ocrImg := gocv.NewMat()
		defer ocrImg.Close()
		gocv.Resize(region, &ocrImg, image.Point{}, 2.0, 2.0, gocv.InterpolationCubic)

		// Save crop for Tesseract
		cropFilename := fmt.Sprintf("web/static/images/snap_crop_%d.jpg", SystemClock())
		gocv.IMWrite(cropFilename, ocrImg)

		// Save full image with box
		filename = fmt.Sprintf("web/static/images/snap_%d.jpg", SystemClock())
		gocv.IMWrite(filename, img)

		// Run OCR on crop
		out, err := exec.Command("tesseract", cropFilename, "stdout", "--psm", "7", "-l", "eng").Output()
		if err == nil {
			detectedText = strings.TrimSpace(string(out))
		} else {
			log.Printf("OCR Failed on crop: %v", err)
		}

		// Clean up crop file to prevent disk fill-up
		defer os.Remove(cropFilename)

	} else {
		// No plate found, save full image and try fallback OCR on full image (risky but better than nothing)
		filename = fmt.Sprintf("web/static/images/snap_%d.jpg", SystemClock())
		gocv.IMWrite(filename, img)

		out, err := exec.Command("tesseract", filename, "stdout", "--psm", "11").Output() // PSM 11: Sparse text
		if err == nil {
			detectedText = strings.TrimSpace(string(out))
		}
	}

	detectedText = cleanPlateText(detectedText)
	if detectedText == "" {
		detectedText = "NOT FOUND"
	}

	return detectedText, filename, nil
}

// processYOLOv8Output parses the output tensor from YOLOv8
// YOLOv8 Output: [Batch, Dimensions, Anchors] -> [1, 5, 8400] for 1 class
// Dimensions: CenterX, CenterY, Width, Height, Score
func processYOLOv8Output(prob gocv.Mat, imgW, imgH int) (image.Rectangle, bool) {
	// Get dimensions
	// sizes := prob.Size()
	// The Mat might be 3D: [1, 5, 8400], but GoCV might return it as 2D [5, 8400] if squeezed?
	// Usually ReadNet returns [1, 5, 8400]
	// We need to access raw data.

	// prob.DataPtrFloat32() returns the flat array.
	// Indexing: [batch][dim][anchor]
	// We assume batch=1.

	// Shapes logic:
	// We need to transpose essentially.
	// Iterate over 8400 anchors.
	// For each anchor, check score (index 4).

	ptr, err := prob.DataPtrFloat32()
	if err != nil {
		log.Println("Error getting data ptr:", err)
		return image.Rectangle{}, false
	}

	totalAnchors := 8400
	numDims := 5 // x, y, w, h, score (assuming 1 class)

	// Check if the output size matches expectation.
	// If the model has more classes, numDims will be higher (4 + num_classes).
	// We determine numDims by dividing total elements by 8400.
	totalElements := prob.Total()
	if totalElements%totalAnchors == 0 {
		numDims = totalElements / totalAnchors
	}

	// YOLOv8 format: [class_prob] is at index 4 onwards.
	// x,y,w,h are 0,1,2,3.

	var bestScore float32 = 0.4 // Threshold
	var bestBox image.Rectangle
	found := false

	// Scale factors (Model is 640x640)
	scaleX := float32(imgW) / 640.0
	scaleY := float32(imgH) / 640.0

	// Iterate columns (anchors)
	for i := 0; i < totalAnchors; i++ {
		// Calculate score.
		// The matrix is [Dimensions, Anchors].
		// So data is laid out: Row 0 (all Xs), Row 1 (all Ys)...
		// Index for attribute A at anchor I is: A * totalAnchors + I

		// Find max class score
		var maxClassScore float32 = 0.0
		// Classes start at index 4
		for c := 4; c < numDims; c++ {
			score := ptr[c*totalAnchors + i]
			if score > maxClassScore {
				maxClassScore = score
			}
		}

		if maxClassScore > bestScore {
			bestScore = maxClassScore
			found = true

			cx := ptr[0*totalAnchors + i]
			cy := ptr[1*totalAnchors + i]
			w := ptr[2*totalAnchors + i]
			h := ptr[3*totalAnchors + i]

			// Convert to corners
			x1 := (cx - w/2) * scaleX
			y1 := (cy - h/2) * scaleY
			x2 := (cx + w/2) * scaleX
			y2 := (cy + h/2) * scaleY

			bestBox = image.Rect(int(x1), int(y1), int(x2), int(y2))
		}
	}

	return bestBox, found
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
	return time.Now().UnixNano()
}
