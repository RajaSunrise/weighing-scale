//go:build gocv

package cv

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

	// Check file size (lowered threshold because some ONNX exports are small or split into .data)
	fileInfo, _ := os.Stat(modelPath)
	if fileInfo.Size() < 100*1024 { // 100KB
		log.Printf("Warning: Model file %s is too small (%d bytes). This appears to be a placeholder. ANPR will be disabled.", modelPath, fileInfo.Size())
		return &ANPRService{IsLoaded: false}
	}

	// Attempt to load the model.
	// Supports YOLOv5 (standard export) and YOLOv8 (onnx export)
	var net gocv.Net
	var loadErr error

	func() {
		defer func() {
			if r := recover(); r != nil {
				loadErr = fmt.Errorf("panic during model load: %v", r)
			}
		}()

		// Convert to absolute path to ensure external data (weights) are found correctly by OpenCV
		// This is crucial for .onnx models that reference external .onnx.data files
		if absPath, err := filepath.Abs(modelPath); err == nil {
			modelPath = absPath
		}

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
		// Enforce TCP for RTSP stability
		if strings.HasPrefix(cameraSource, "rtsp") {
			os.Setenv("OPENCV_FFMPEG_CAPTURE_OPTIONS", "rtsp_transport;tcp")
		}
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

	// Parse YOLO Output (supports v5 and v8 shapes)
	bestBox, found := processYOLOOutput(prob, img.Cols(), img.Rows())

	var filename string
	var detectedText string

	if found {
		// Crop the license plate
		rect := bestBox
		// Bounds check
		if rect.Min.X < 0 { rect.Min.X = 0 }
		if rect.Min.Y < 0 { rect.Min.Y = 0 }
		if rect.Max.X > img.Cols() { rect.Max.X = img.Cols() }
		if rect.Max.Y > img.Rows() { rect.Max.Y = img.Rows() }

		gocv.Rectangle(&img, rect, color.RGBA{0, 255, 0, 0}, 2)

		region := img.Region(rect)

		ocrImg := gocv.NewMat()
		defer ocrImg.Close()
		gocv.Resize(region, &ocrImg, image.Point{}, 2.0, 2.0, gocv.InterpolationCubic)
		region.Close() // Close region explicitly

		cropFilename := fmt.Sprintf("web/static/images/snap_crop_%d.jpg", SystemClock())
		gocv.IMWrite(cropFilename, ocrImg)

		filename = fmt.Sprintf("web/static/images/snap_%d.jpg", SystemClock())
		gocv.IMWrite(filename, img)

		// Run OCR on crop
		out, err := exec.Command("tesseract", cropFilename, "stdout", "--psm", "7", "-l", "eng").Output()
		if err == nil {
			detectedText = strings.TrimSpace(string(out))
		} else {
			log.Printf("OCR Failed on crop: %v", err)
		}

		defer os.Remove(cropFilename)

	} else {
		filename = fmt.Sprintf("web/static/images/snap_%d.jpg", SystemClock())
		gocv.IMWrite(filename, img)

		out, err := exec.Command("tesseract", filename, "stdout", "--psm", "11").Output()
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

// processYOLOOutput parses output tensor from either YOLOv5 or YOLOv8
func processYOLOOutput(prob gocv.Mat, imgW, imgH int) (image.Rectangle, bool) {
	ptr, err := prob.DataPtrFloat32()
	if err != nil {
		log.Println("Error getting data ptr:", err)
		return image.Rectangle{}, false
	}

	totalElements := prob.Total()

	// Determine shape structure
	// We expect 3 dimensions usually: [Batch, Dimensions, Anchors] (v8) or [Batch, Anchors, Dimensions] (v5)
	// But GoCV might flatten/squeeze.
	// Common Anchor counts: 25200 (v5 640x640), 8400 (v8 640x640)

	// Heuristic: Check common anchor counts
	isTransposed := false // v8 style: [dims][anchors]
	numAnchors := 0
	numDims := 0

	if totalElements % 8400 == 0 {
		numAnchors = 8400
		numDims = totalElements / 8400
		isTransposed = true // v8 standard
	} else if totalElements % 25200 == 0 {
		numAnchors = 25200
		numDims = totalElements / 25200
		isTransposed = false // v5 standard
	} else {
		// Fallback/Unknown - Assume v8 style 8400 if close, otherwise fail safe
		// Or try to infer from shape if accessible (gocv Mat size is int[])
		size := prob.Size()
		if len(size) >= 3 {
			// [1, 5, 8400]
			if size[2] > size[1] {
				numAnchors = size[2]
				numDims = size[1]
				isTransposed = true
			} else {
				numAnchors = size[1]
				numDims = size[2]
				isTransposed = false
			}
		} else {
			return image.Rectangle{}, false
		}
	}

	// Sanity check dimensions (at least x,y,w,h,conf)
	if numDims < 5 {
		return image.Rectangle{}, false
	}

	var bestScore float32 = 0.4
	var bestBox image.Rectangle
	found := false

	scaleX := float32(imgW) / 640.0
	scaleY := float32(imgH) / 640.0

	for i := 0; i < numAnchors; i++ {
		var cx, cy, w, h, score float32

		if isTransposed {
			// [Dims, Anchors]
			// Index = Dim * numAnchors + i

			// Find max class score
			var maxClassScore float32 = 0.0
			for c := 4; c < numDims; c++ {
				val := ptr[c*numAnchors + i]
				if val > maxClassScore {
					maxClassScore = val
				}
			}

			score = maxClassScore // v8 usually combines obj_conf * class_conf, or just class_conf

			if score > bestScore {
				cx = ptr[0*numAnchors + i]
				cy = ptr[1*numAnchors + i]
				w  = ptr[2*numAnchors + i]
				h  = ptr[3*numAnchors + i]
			}
		} else {
			// [Anchors, Dims] (v5 standard)
			// Index = i * numDims + Dim

			objConf := ptr[i*numDims + 4]
			if objConf < bestScore {
				continue
			}

			var maxClassScore float32 = 0.0
			for c := 5; c < numDims; c++ {
				val := ptr[i*numDims + c]
				if val > maxClassScore {
					maxClassScore = val
				}
			}

			score = objConf * maxClassScore

			if score > bestScore {
				cx = ptr[i*numDims + 0]
				cy = ptr[i*numDims + 1]
				w  = ptr[i*numDims + 2]
				h  = ptr[i*numDims + 3]
			}
		}

		if score > bestScore {
			bestScore = score
			found = true

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
