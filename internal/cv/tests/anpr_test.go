package tests

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"stoneweigh/internal/cv"
)

// Helper to download images
func downloadImage(url, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func TestANPRDetection(t *testing.T) {
	// URLs provided by user
	urls := []string{
		"https://imgs.search.brave.com/vFZSqoGsFA12xPOK12HHR7qcubUUZ4ZuVYlDoLhRGf8/rs:fit:860:0:0:0/g:ce/aHR0cHM6Ly9yYWRh/cm11a29tdWtvLmRp/c3dheS5pZC91cGxv/YWQvY2U4YmRmOGI0/YWE3ZjgxZTU4NjVk/MDc4MTVhNDk3NTgu/anBn",
		"https://imgs.search.brave.com/WAiIZ3JutY27HMPQowl4TNgbqp_N1Vgo3r580yw7GQA/rs:fit:860:0:0:0/g:ce/aHR0cHM6Ly9pbWd4/LmdyaWRvdG8uY29t/L2Nyb3AvMHgwOjB4/MC83MDB4NDY1L2Zp/bHRlcnM6d2F0ZXJt/YXJrKGZpbGUvMjAx/Ny9ncmlkb3RvL2lt/Zy93YXRlcm1hcmtf/b3Rvc2VrZW4ucG5n/LDUsNSw2MCkvcGhv/dG8vMjAyMC8wNi8w/My82MTYyNjkzODMu/anBlZw",
		"https://imgs.search.brave.com/3eBbBInioYUx97vfutKvMvRWTmXl2wBoNXuoUcByMWs/rs:fit:860:0:0:0/g:ce/aHR0cHM6Ly9zdGF0/aWsudGVtcG8uY28v/ZGF0YS8yMDI0LzA2/LzA3L2lkXzEzMDg1/NTAvMTMwODU1MF83/MjAuanBn",
	}

	// Expected results map (key: index + 1)
	expectedPlates := map[int]string{
		1: "B 8187",
		2: "B 9190 IC",
		3: "K 8324 QD",
	}

	// Initialize Service
	// Path assumes running from repo root via `go test ./internal/cv/tests/...`.
	// But `models/` is at repo root. `internal/cv/tests` is 3 dirs deep relative to root?
	// Actually, `go test` sets the working directory to the package directory.
	// So we are in `internal/cv/tests`.
	// To reach root: `../../../`
	modelPath := "../../../models/platdetection.onnx"

	// Check if we are running from root (e.g. go test ./...) or from package.
	// Best is to use absolute path or try both.
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		// Try from root
		if _, err := os.Stat("models/platdetection.onnx"); err == nil {
			modelPath = "models/platdetection.onnx"
		}
	}

	anpr := cv.NewANPRService(modelPath)

	if !anpr.IsLoaded {
		fmt.Println("WARNING: ANPR Model not loaded (Mock Mode or Missing File). Detection will be simulated.")
	}

	for i, url := range urls {
		idx := i + 1
		filename := fmt.Sprintf("test_image_%d.jpg", idx)
		fmt.Printf("Downloading Image %d...\n", idx)

		err := downloadImage(url, filename)
		if err != nil {
			t.Errorf("Failed to download image %d: %v", idx, err)
			continue
		}
		defer os.Remove(filename) // Cleanup

		fmt.Printf("Processing Image %d...\n", idx)

		plate, snapshot, err := anpr.CaptureAndDetect(filename)
		if err != nil {
			t.Errorf("Error detecting image %d: %v", idx, err)
		} else {
			fmt.Printf("RESULT Image %d: Detected Plate: %s\n", idx, plate)
			fmt.Printf("       Snapshot saved to: %s\n", snapshot)

			// Assertion
			expected := expectedPlates[idx]
			if !strings.Contains(plate, expected) {
				t.Errorf("Image %d: Expected plate to contain '%s', got '%s'", idx, expected, plate)
			}
			fmt.Println("------------------------------------------------")
		}
	}
}
