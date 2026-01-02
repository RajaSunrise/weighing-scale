package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RequestLogger logs the details of each request
var (
	lastHealthLog time.Time
	healthLogMu   sync.Mutex
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		path := c.Request.URL.Path

		// Logic specifically for /health to log only once every 6 hours (unless error)
		if path == "/health" {
			if c.Writer.Status() == http.StatusOK {
				healthLogMu.Lock()
				if time.Since(lastHealthLog) < 6*time.Hour {
					healthLogMu.Unlock()
					return
				}
				lastHealthLog = time.Now()
				healthLogMu.Unlock()
			}
		}

		latency := time.Since(startTime)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		log.Printf("[HTTP] %3d | %13v | %15s | %-7s %s",
			status,
			latency,
			clientIP,
			method,
			path,
		)
	}
}

// AuthRequired checks if the user is logged in
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user_id")
		if user == nil {
			// Check if it's an API call or HTML request
			if c.GetHeader("Accept") == "application/json" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			} else {
				c.Redirect(http.StatusFound, "/login")
				c.Abort()
			}
			return
		}
		c.Next()
	}
}

// RoleRequired checks for specific roles
func RoleRequired(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userRole := session.Get("role")
		if userRole != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.Next()
	}
}

// RateLimiter implements a simple IP-based rate limiter using token bucket
func RateLimiter(limit rate.Limit, burst int) gin.HandlerFunc {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Background cleanup for old entries
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(limit, burst)}
		}
		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
			return
		}
		mu.Unlock()
		c.Next()
	}
}
