package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func BenchmarkRateLimiter(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(RateLimiter(context.Background(), rate.Limit(1000000), 1000000))
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(w, req)
	}
}
