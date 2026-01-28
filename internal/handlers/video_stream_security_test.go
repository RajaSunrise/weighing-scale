//go:build !gocv

package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
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

func setupVideoServer(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:memdb_video%d?mode=memory&cache=shared", time.Now().UnixNano())
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

func TestProxyVideo_RateLimit(t *testing.T) {
	// 0. Mock FFmpeg
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "mock_bin")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a dummy ffmpeg script that sleeps
	ffmpegPath := filepath.Join(tmpDir, "ffmpeg")
	// Use a script that sleeps and outputs something so Start() succeeds
	script := `#!/bin/sh
# Sleep to simulate streaming
sleep 2
echo "done"
`
	err = os.WriteFile(ffmpegPath, []byte(script), 0755)
	assert.NoError(t, err)

	// Add tmpDir to PATH
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	r, db := setupVideoServer(t)

	// 1. Establish Session
	w_login := httptest.NewRecorder()
	req_login, _ := http.NewRequest("POST", "/login_mock_video", nil)
	r.ServeHTTP(w_login, req_login)
	cookieVal := w_login.Header().Get("Set-Cookie")

	// 2. Setup "Hanging" Server
	// We don't technically need the hanging server if our mock ffmpeg hangs (sleeps).
	// But ProxyVideo connects to URL *via ffmpeg*.
	// Our mock ffmpeg ignores arguments and just sleeps.
	// So we can pass any URL.

	// 3. Seed Camera
	station := models.WeighingStation{Name: "S1", Enabled: true}
	db.Create(&station)
	cam := models.StationCamera{Name: "C1", RTSPURL: "http://example.com/dummy", WeighingStationID: station.ID}
	db.Create(&cam)

	// 4. Launch concurrent requests
	var wg sync.WaitGroup
	results := make(chan int, 20)

	// Launch 10 requests. Limit should be 5.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			// Use context with timeout to force return of blocking requests
			// Timeout should be LESS than the sleep in mock ffmpeg (2s)
			// But enough for all requests to start (500ms).
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			req, _ := http.NewRequestWithContext(ctx, "GET", "/api/camera/stream?camera_id=1", nil)
			req.Header.Set("Cookie", cookieVal)
			r.ServeHTTP(w, req)
			results <- w.Code
		}()
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()
	close(results)

	statusCodes := make(map[int]int)
	for code := range results {
		statusCodes[code]++
	}

	t.Logf("Status Codes: %v", statusCodes)

	// Verification
	// The mock ffmpeg sleeps for 2s.
	// The requests timeout after 500ms.
	// So successful starts will return 200 (headers written) then timeout.
	// If rate limiting works, some will get 429 immediately.

	if statusCodes[http.StatusTooManyRequests] == 0 {
		t.Error("Expected 429 Too Many Requests responses, got none")
	}
}
