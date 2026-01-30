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

	"stoneweigh/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Fallback to FFmpeg if GoCV is not available
// This allows the system to stream RTSP even without CGO/OpenCV bindings,
// as long as the 'ffmpeg' binary is in the PATH.

// ProxyVideo handles the RTSP to MJPEG conversion using FFmpeg
func (s *Server) ProxyVideo(c *gin.Context) {
	// SECURITY FIX: Prevent SSRF by strictly requiring camera_id or station_id lookup
	// Do NOT accept raw 'url' parameter from client.
	camID := c.Query("camera_id")
	stationID := c.Query("station_id")

	if camID == "" && stationID == "" {
		c.String(http.StatusBadRequest, "Missing camera_id or station_id")
		return
	}

	// Lookup Camera URL from Database
	var url string
	var targetStationID uint

	if camID != "" {
		// Priority 1: Specific Camera ID
		var cam models.StationCamera
		if err := s.DB.First(&cam, camID).Error; err == nil {
			url = cam.RTSPURL
			targetStationID = cam.WeighingStationID
		}
	} else if stationID != "" {
		// Priority 2: Station ID (Legacy / Default Camera)
		var station models.WeighingStation
		if err := s.DB.Preload("Cameras").First(&station, stationID).Error; err == nil {
			targetStationID = station.ID
			if len(station.Cameras) > 0 {
				url = station.Cameras[0].RTSPURL
			} else {
				url = station.CameraURL
			}
		}
	}

	if url == "" {
		c.String(http.StatusNotFound, "Camera not found or invalid ID")
		return
	}

	// Verify User Access (Defense in Depth)
	// Users should only see cameras for stations they are assigned to.
	session := sessions.Default(c)
	role := session.Get("role")
	userID := session.Get("user_id")

	if role != "admin" {
		var assignment models.UserStationAssignment
		if err := s.DB.Where("user_id = ? AND weighing_station_id = ?", userID, targetStationID).First(&assignment).Error; err != nil {
			c.String(http.StatusForbidden, "Access denied to this camera")
			return
		}
	}

	// Rate Limit: Prevent DoS via process exhaustion
	select {
	case streamSemaphore <- struct{}{}:
		defer func() { <-streamSemaphore }()
	default:
		c.String(http.StatusTooManyRequests, "Too many active streams")
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
	// Optimized split function to find JPEG boundaries
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// Find SOI (FF D8)
		soi := bytes.Index(data, []byte{0xFF, 0xD8})
		if soi == -1 {
			// Optimization: If no SOI found, discard all but the last byte
			// (in case the last byte is 0xFF and next is 0xD8)
			if len(data) > 1 {
				return len(data) - 1, nil, nil
			}
			return 0, nil, nil
		}

		// Optimization: If SOI is not at the start, discard data before it
		if soi > 0 {
			return soi, nil, nil
		}

		// Find EOI (FF D9) after SOI (starts at 0)
		if len(data) < 2 {
			return 0, nil, nil
		}

		eoi := bytes.Index(data[2:], []byte{0xFF, 0xD9})
		if eoi == -1 {
			return 0, nil, nil
		}

		// eoi is relative to data[2:], so actual index is 2 + eoi
		end := 2 + eoi + 2
		return end, data[0:end], nil
	}

	scanner.Split(split)

	// Increase buffer for high-res images
	buf := make([]byte, 1024*1024)   // 1MB
	scanner.Buffer(buf, 5*1024*1024) // 5MB max

	for scanner.Scan() {
		select {
		case <-c.Request.Context().Done():
			return
		default:
			frame := scanner.Bytes()

			// Write boundary
			_, err := c.Writer.Write(fmt.Appendf(nil, "--frame\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(frame)))
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
	streamMap       = make(map[string]*any)
	streamLock      sync.Mutex
	streamSemaphore = make(chan struct{}, 5) // Limit concurrent FFmpeg processes
)
