//go:build gocv

package handlers

import (
	"fmt"
	"image"
	"net/http"
	"strings"
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
	ticker := time.NewTicker(25 * time.Millisecond) // 25 FPS
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
	img := gocv.NewMat()
	defer img.Close()

	for {
		// Outer loop for reconnection
		select {
		case <-s.Stop:
			return
		default:
		}

		// Force FFMPEG backend to avoid GStreamer frame estimation warnings
		// Also use TCP for RTSP to prevent UDP timeout warnings
		// Note: The OPENCV_FFMPEG_CAPTURE_OPTIONS env var is set globally in main.go to avoid race conditions here.
		// However, if we need per-stream options in the future, we would need a different approach (e.g. videoio properties).
		
		// Prepare connection URL
		connectUrl := s.URL
		if strings.HasPrefix(connectUrl, "rtsp") {
			// Append rtsp_transport=tcp to the URL parameters to enforce TCP at the source level for FFmpeg.
			// This is often more reliable than the env var alone for specific streams.
			separator := "?"
			if strings.Contains(connectUrl, "?") {
				separator = "&"
			}
			if !strings.Contains(connectUrl, "rtsp_transport") {
				connectUrl = connectUrl + separator + "rtsp_transport=tcp"
			}
		}

		// Use auto-detection instead of enforcing FFmpeg backend ID (1900).
		// Enforcing the backend sometimes causes "backend is generally available but can't be used to capture by name"
		// error if the build configuration or URL scheme conflicts.
		// Auto-detection will still prefer FFmpeg for RTSP in most standard OpenCV builds,
		// and will respect the OPENCV_FFMPEG_CAPTURE_OPTIONS env var if FFmpeg is selected.
		vc, err := gocv.OpenVideoCapture(connectUrl)
		if err != nil {
			fmt.Printf("Error opening stream %s: %v\n", s.URL, err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Optimize buffer size for low latency
		vc.Set(gocv.VideoCaptureBufferSize, 1)

		// Inner loop for reading frames
		for {
			select {
			case <-s.Stop:
				vc.Close()
				return
			default:
				if ok := vc.Read(&img); !ok || img.Empty() {
					// Stream disconnected or empty frame
					time.Sleep(100 * time.Millisecond)
					vc.Close()
					goto Reconnect
				}

				// Resize to 480p (854x480)
				gocv.Resize(img, &img, image.Point{X: 854, Y: 480}, 0, 0, gocv.InterpolationLinear)

				// Encode to JPG with reduced quality (70) to save bandwidth
				buf, err := gocv.IMEncodeWithParams(gocv.JPEGFileExt, img, []int{gocv.IMWriteJpegQuality, 70})
				if err == nil {
					// CRITICAL FIX: Copy data to Go memory
					// buf.GetBytes() returns a slice backed by C++ memory which is freed on buf.Close()
					data := buf.GetBytes()
					dst := make([]byte, len(data))
					copy(dst, data)

					s.LastFrameMu.Lock()
					s.LastFrame = dst
					s.LastFrameMu.Unlock()
					buf.Close()
				}

				// Cap framerate
				time.Sleep(1 * time.Millisecond)
			}
		}

	Reconnect:
		// Break out of inner loop to reconnect
		continue
	}
}
