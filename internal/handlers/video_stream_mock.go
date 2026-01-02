//go:build !gocv

package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// ProxyVideo handles the RTSP to MJPEG conversion (Mock for non-GoCV builds)
func (s *Server) ProxyVideo(c *gin.Context) {
	c.String(http.StatusNotImplemented, "Video streaming requires OpenCV (gocv build tag)")
}
