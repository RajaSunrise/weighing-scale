//go:build gocv

package handlers

import (
	"fmt"
	"image"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gocv.io/x/gocv"
)

// StreamMap holds active streams to prevent opening too many connections to the same camera
var (
	streamMap  = make(map[string]*SharedStream)
	streamLock sync.Mutex
)

type SharedStream struct {
	URL         string
	Clients     int
	Broadcast   chan []byte
	Stop        chan bool
	LastFrame   []byte
	LastFrameMu sync.RWMutex
}

// ProxyVideo handles the RTSP to MJPEG conversion
func (s *Server) ProxyVideo(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		// Try to find by camera ID if provided
		camID := c.Query("camera_id")
		if camID != "" {
			// Find camera in DB
			// ... logic to lookup models.StationCamera
			// For now, we assume direct URL passed or handled by frontend
		}
	}

	if url == "" {
		c.String(http.StatusBadRequest, "Missing URL")
		return
	}

	// Set headers for MJPEG
	c.Writer.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	// Get or Create Stream
	stream := getStream(url)
	streamLock.Lock()
	stream.Clients++
	streamLock.Unlock()

	defer func() {
		streamLock.Lock()
		stream.Clients--
		if stream.Clients <= 0 {
			close(stream.Stop)
			delete(streamMap, url)
		}
		streamLock.Unlock()
	}()

	// Stream Loop
	ticker := time.NewTicker(100 * time.Millisecond) // 10 FPS
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			stream.LastFrameMu.RLock()
			frame := stream.LastFrame
			stream.LastFrameMu.RUnlock()

			if len(frame) == 0 {
				continue
			}

			// Write MIME boundary
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
}

func getStream(url string) *SharedStream {
	streamLock.Lock()
	defer streamLock.Unlock()

	if s, ok := streamMap[url]; ok {
		return s
	}

	s := &SharedStream{
		URL:       url,
		Stop:      make(chan bool),
		Broadcast: make(chan []byte),
	}
	streamMap[url] = s

	go captureLoop(s)
	return s
}

func captureLoop(s *SharedStream) {
	vc, err := gocv.OpenVideoCapture(s.URL)
	if err != nil {
		fmt.Printf("Error opening stream %s: %v\n", s.URL, err)
		return
	}
	defer vc.Close()

	img := gocv.NewMat()
	defer img.Close()

	for {
		select {
		case <-s.Stop:
			return
		default:
			if ok := vc.Read(&img); !ok || img.Empty() {
				// Reconnect logic or sleep
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Resize to reduce bandwidth
			gocv.Resize(img, &img, image.Point{X: 640, Y: 360}, 0, 0, gocv.InterpolationLinear)

			// Encode to JPG
			buf, err := gocv.IMEncode(".jpg", img)
			if err == nil {
				s.LastFrameMu.Lock()
				s.LastFrame = buf.GetBytes()
				s.LastFrameMu.Unlock()
				buf.Close()
			}

			// Cap framerate
			time.Sleep(50 * time.Millisecond) // ~20 FPS max
		}
	}
}
