//go:build !gocv

package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"stoneweigh/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupVideoServerVerify(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:memdb_video_verify%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(
		&models.User{},
		&models.WeighingStation{},
		&models.StationCamera{},
		&models.UserStationAssignment{},
	)
	assert.NoError(t, err)

	server := NewServer(db, nil, nil, nil)

	r := gin.New()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	r.GET("/api/camera/stream", server.ProxyVideo)

	r.POST("/login_mock_video", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("user_id", uint(1))
		session.Set("username", "Admin")
		session.Set("role", "admin")
		session.Save()
		c.Status(200)
	})

	return r, db
}

func TestProxyVideo_ContentVerification(t *testing.T) {
	// 0. Mock FFmpeg
	tmpDir, err := os.MkdirTemp("", "mock_bin_verify")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	ffmpegPath := filepath.Join(tmpDir, "ffmpeg")
	// Output two JPEG frames: FF D8 01 FF D9 and FF D8 02 FF D9
	// Use octal codes for printf compatibility
	script := `#!/bin/sh
printf "\377\330\001\377\331"
sleep 0.1
printf "\377\330\002\377\331"
`
	err = os.WriteFile(ffmpegPath, []byte(script), 0755)
	assert.NoError(t, err)

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	r, db := setupVideoServerVerify(t)

	// Login
	w_login := httptest.NewRecorder()
	req_login, _ := http.NewRequest("POST", "/login_mock_video", nil)
	r.ServeHTTP(w_login, req_login)
	cookieVal := w_login.Header().Get("Set-Cookie")

	// Seed Camera
	station := models.WeighingStation{Name: "S1", Enabled: true}
	db.Create(&station)
	cam := models.StationCamera{Name: "C1", RTSPURL: "http://example.com/dummy", WeighingStationID: station.ID}
	db.Create(&cam)

	// Request Stream
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/camera/stream?camera_id=1", nil)
	req.Header.Set("Cookie", cookieVal)

	// We need to run the server in a way that we can read the stream.
	// httptest.Recorder buffers everything, which is fine for small output.
	r.ServeHTTP(w, req)

	// Verify Headers
	assert.Equal(t, http.StatusOK, w.Code)
	contentType := w.Header().Get("Content-Type")
	assert.Contains(t, contentType, "multipart/x-mixed-replace")
	assert.Contains(t, contentType, "boundary=frame")

	// Verify Body
	// We expect two frames.
	body := w.Body.Bytes()
	if len(body) == 0 {
		t.Fatal("Response body is empty")
	}

	// Verify Body contains the frames
	// Since the stream might end abruptly (MJPEG doesn't enforce a closing boundary),
	// multipart.Reader might return Unexpected EOF.
	// We manually verify the content exists in the body.

	frame1 := []byte{0xFF, 0xD8, 0x01, 0xFF, 0xD9}
	frame2 := []byte{0xFF, 0xD8, 0x02, 0xFF, 0xD9}

	if !bytes.Contains(body, frame1) {
		t.Errorf("Body does not contain frame 1")
	}
	if !bytes.Contains(body, frame2) {
		t.Errorf("Body does not contain frame 2")
	}

	// Also verify boundary exists
	if !bytes.Contains(body, []byte("--frame")) {
		t.Errorf("Body does not contain boundary")
	}
}
