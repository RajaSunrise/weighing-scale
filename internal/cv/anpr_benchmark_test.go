//go:build gocv

package cv

import (
	"image"
	"image/color"
	"os"
	"testing"

	"gocv.io/x/gocv"
)

// BenchmarkCaptureAndDetect measures the performance of the ANPR pipeline.
// Use: go test -bench=. -tags gocv ./internal/cv/...
func BenchmarkCaptureAndDetect(b *testing.B) {
	// Attempt to initialize ANPR Service
	// Adjust path to model if necessary relative to test execution
	anpr := NewANPRService("../../models/platdetection.onnx")
	if !anpr.IsLoaded {
		b.Skip("ANPR Model not loaded or invalid, skipping benchmark")
	}

	// Create a dummy image that resembles a vehicle with a plate
	// In a real scenario, use a real image.
	img := gocv.NewMatWithSize(640, 640, gocv.MatTypeCV8UC3)
	defer img.Close()

	// Draw a rectangle and some text to simulate a plate (roughly)
	// This might not trigger YOLO detection, so it might hit the fallback path (Full Image OCR)
	gocv.Rectangle(&img, image.Rect(200, 200, 440, 300), color.RGBA{255, 255, 255, 0}, -1)
	gocv.PutText(&img, "B 1234 XY", image.Pt(220, 270), gocv.FontHersheySimplex, 1.5, color.RGBA{0, 0, 0, 0}, 3)

	// Save to temporary file
	testFile := "benchmark_test_img.jpg"
	gocv.IMWrite(testFile, img)
	defer os.Remove(testFile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// We expect this to process the image.
		// Since OpenVideoCapture works with image files as single-frame videos usually:
		_, _, err := anpr.CaptureAndDetect(testFile)
		if err != nil {
			b.Errorf("CaptureAndDetect failed: %v", err)
		}
	}
}
