package tests

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
		"https://imgs.search.brave.com/dcHqTSwN72unv7Xy1pCmws_ePw0cjWwRIslT8se4zmk/rs:fit:860:0:0:0/g:ce/aHR0cHM6Ly9zLmdh/cmFzaS5pZC9jMTIw/MHg2NzUvcTk5L2Fy/dGljbGUvZTFjNzJl/YzItMTQ5ZS00MzM2/LWE5Y2UtNzQ2YzNj/MTMyNGMxLmpwZWc",
		"https://imgs.search.brave.com/WAiIZ3JutY27HMPQowl4TNgbqp_N1Vgo3r580yw7GQA/rs:fit:860:0:0:0/g:ce/aHR0cHM6Ly9pbWd4/LmdyaWRvdG8uY29t/L2Nyb3AvMHgwOjB4/MC83MDB4NDY1L2Zp/bHRlcnM6d2F0ZXJt/YXJrKGZpbGUvMjAx/Ny9ncmlkb3RvL2lt/Zy93YXRlcm1hcmtf/b3Rvc2VrZW4ucG5n/LDUsNSw2MCkvcGhv/dG8vMjAyMC8wNi8w/My82MTYyNjkzODMu/anBlZw",
		"https://imgs.search.brave.com/0pXu9kATZ_aeNvwaFRZqxCwdpV3FfOUN18ePtMwhVgo/rs:fit:860:0:0:0/g:ce/aHR0cHM6Ly9iZXJ0/dWFocG9zLmNvbS93/cC1jb250ZW50L3Vw/bG9hZHMvMjAyNS8w/Ny9QbGF0LU5vbi1C/TS03NTB4Njc1Lmpw/ZWc",
	}

	// Initialize Service
	// Path assumes running from repo root. In test context, might differ.
	// But let's assume standard go test execution.
	anpr := cv.NewANPRService("../../models/platdetection.pt")

	if !anpr.IsLoaded {
		fmt.Println("WARNING: ANPR Model not loaded (Mock Mode or Missing File). Detection will be simulated.")
	}

	for i, url := range urls {
		filename := fmt.Sprintf("test_image_%d.jpg", i+1)
		fmt.Printf("Downloading Image %d...\n", i+1)

		err := downloadImage(url, filename)
		if err != nil {
			t.Errorf("Failed to download image %d: %v", i+1, err)
			continue
		}
		defer os.Remove(filename) // Cleanup

		fmt.Printf("Processing Image %d...\n", i+1)

		// CaptureAndDetect expects a camera URL/ID, but under the hood GoCV opens it.
		// GoCV OpenVideoCapture works with files too!
		// So we pass the filename.

		plate, snapshot, err := anpr.CaptureAndDetect(filename)
		if err != nil {
			t.Errorf("Error detecting image %d: %v", i+1, err)
		} else {
			fmt.Printf("RESULT Image %d: Detected Plate: %s\n", i+1, plate)
			fmt.Printf("       Snapshot saved to: %s\n", snapshot)
			fmt.Println("------------------------------------------------")
		}
	}
}
