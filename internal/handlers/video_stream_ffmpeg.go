//go:build !gocv

package handlers

import (
	"bufio"
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"os/exec"
	"sync"

	"github.com/gin-gonic/gin"
)

// Fallback to FFmpeg if GoCV is not available
// This allows the system to stream RTSP even without CGO/OpenCV bindings,
// as long as the 'ffmpeg' binary is in the PATH.

// ProxyVideo handles the RTSP to MJPEG conversion using FFmpeg
func (s *Server) ProxyVideo(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.String(http.StatusBadRequest, "Missing URL")
		return
	}

	// Set headers for MJPEG
	c.Writer.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	// Start FFmpeg
	// -i url: Input
	// -f image2pipe: Output format suitable for piping
	// -vcodec mjpeg: Output MJPEG
	// -q:v 5: Quality (1-31, lower is better)
	// -r 5: Frame rate (low to save bandwidth)
	// -: Output to stdout

	// Note: We use -rtsp_transport tcp to be more robust over internet
	cmd := exec.Command("ffmpeg",
		"-rtsp_transport", "tcp",
		"-i", url,
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"-q:v", "5",
		"-r", "5",
		"-")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("FFmpeg stdout error:", err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("FFmpeg start error:", err)
		c.String(http.StatusInternalServerError, "Failed to start ffmpeg")
		return
	}

	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait() // Prevent zombie process
		}
	}()

	// Read loop
	// We need to parse JPEG chunks.
	// JPEG starts with FF D8, ends with FF D9.

	// Using a buffered reader
	reader := bufio.NewReader(stdout)

	// Naive approach: Read until EOF.
	// But we need to frame it for multipart.
	// Actually, image2pipe outputs a stream of concatenated JPEGs.
	// We can scan for FFD8 ... FFD9.

	// Create a multipart writer wrapping the response writer?
	// No, we need to write boundaries manually to flush correctly.

	mw := multipart.NewWriter(c.Writer)
	mw.SetBoundary("frame")

	// State machine: 0=Searching SOI, 1=Reading Data
	// Optimally, we read 4KB chunks and search.

	// Simpler: Just read byte by byte? Too slow in Go?
	// Probably OK for 5fps.

	// Better: Use Scanner with custom split?
	// Split on FFD9 (EOI).

	scanner := bufio.NewScanner(reader)
	// Split function to find JPEG boundaries
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// Find SOI (FF D8)
		soi := bytes.Index(data, []byte{0xFF, 0xD8})
		if soi == -1 {
			// No SOI found, request more data
			return 0, nil, nil
		}

		// Find EOI (FF D9) after SOI
		eoi := bytes.Index(data[soi:], []byte{0xFF, 0xD9})
		if eoi == -1 {
			// No EOI found, request more data
			return 0, nil, nil
		}

		// Full JPEG found
		end := soi + eoi + 2
		return end, data[soi:end], nil
	}

	scanner.Split(split)

	// Increase buffer for high-res images
	buf := make([]byte, 1024*1024) // 1MB
	scanner.Buffer(buf, 5*1024*1024) // 5MB max

	for scanner.Scan() {
		select {
		case <-c.Request.Context().Done():
			return
		default:
			frame := scanner.Bytes()

			// Write boundary
			_, err := c.Writer.Write([]byte(fmt.Sprintf("--frame\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(frame))))
			if err != nil {
				return
			}
			_, err = c.Writer.Write(frame)
			if err != nil {
				return
			}
			_, err = c.Writer.Write([]byte("\r\n"))
			if err != nil {
				return
			}
			c.Writer.Flush()
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("FFmpeg scan error:", err)
	}
}

// Stub for shared streams if needed, but for fallback we just spawn one process per request
// to keep it simple and stateless.
var (
	streamMap  = make(map[string]*interface{})
	streamLock sync.Mutex
)
