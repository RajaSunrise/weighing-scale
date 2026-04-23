package router

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	csrf "github.com/utrack/gin-csrf"
	"golang.org/x/time/rate"

	"stoneweigh/internal/api"
	"stoneweigh/internal/handlers"
	"stoneweigh/internal/middleware"
	"stoneweigh/internal/pkg/templates"
)

func SetupRouter(ctx context.Context, server *handlers.Server) *gin.Engine {
	// 1. Initialize Gin
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health"},
	}))
	r.Use(gin.Recovery())

	// 2. Setup Session Store
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		if gin.Mode() == gin.ReleaseMode {
			log.Fatal("SESSION_SECRET environment variable is required in production mode")
		}
		// Generate random 32 byte hex string for development/fallback
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			panic("failed to generate random session secret: " + err.Error())
		}
		secret = hex.EncodeToString(b)
		log.Println("WARNING: SESSION_SECRET not set. Using generated random secret. Sessions will not persist across restarts.")
	}
	store := cookie.NewStore([]byte(secret))
	secureCookies := os.Getenv("SECURE_COOKIES") == "true"
	store.Options(sessions.Options{
		MaxAge:   30 * 24 * 60 * 60, // 30 days
		Path:     "/",
		Secure:   secureCookies,        // Dynamic based on env
		HttpOnly: true,                 // Security: prevent XSS
		SameSite: http.SameSiteLaxMode, // Allow same-site cookies (Lax provides CSRF protection for GET)
	})
	r.Use(sessions.Sessions("stoneweigh_session", store))

	// 3. Global Middleware
	r.Use(middleware.RequestLogger())
	r.Use(middleware.SecurityHeaders()) // Add Security Headers
	// Rate Limit: 20 requests/second, burst of 50
	r.Use(middleware.RateLimiter(ctx, rate.Limit(20), 50))

	// CSRF Protection (Skipping /api/external)
	r.Use(func(c *gin.Context) {
		// Skip CSRF for external API
		// SECURITY FIX: Normalize path using path.Clean to prevent bypass with traversal or multiple slashes
		cleanPath := path.Clean(c.Request.URL.Path)
		if cleanPath == "/api/external" || strings.HasPrefix(cleanPath, "/api/external/") {
			c.Next()
			return
		}
		csrf.Middleware(csrf.Options{
			Secret: secret,
			ErrorFunc: func(c *gin.Context) {
				c.String(400, "CSRF token mismatch")
				c.Abort()
			},
		})(c)
	})

	// 4. Static Files & Templates
	r.SetFuncMap(template.FuncMap{
		"json": func(v any) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a)
		},
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"currentYear": func() int {
			return time.Now().Year()
		},
	})

	// Cache Control for Static Files (1 day for public assets, private for reports/snaps)
	r.Use(func(c *gin.Context) {
		p := c.Request.URL.Path
		if strings.HasPrefix(p, "/static") {
			// SECURITY FIX: Prevent public caching of sensitive reports and snapshots
			if strings.HasPrefix(p, "/static/reports/") || strings.Contains(p, "/snap_") {
				c.Header("Cache-Control", "private, no-store, must-revalidate")
				c.Header("Pragma", "no-cache")
				c.Header("Expires", "0")
			} else {
				c.Header("Cache-Control", "public, max-age=86400")
			}
		}
		c.Next()
	})

	r.Static("/static", "./web/static")
	// Fix for legacy/broken links pointing to /web/static
	r.Static("/web/static", "./web/static")

	// Load templates recursively using custom helper
	if err := templates.LoadTemplates(r, "web/templates"); err != nil {
		panic("Failed to load templates: " + err.Error())
	}

	// 5. Public Routes (Site)
	r.GET("/", server.ShowHome)
	r.GET("/produk", server.ShowProduct)
	r.GET("/galeri", server.ShowGallery)
	r.GET("/tentang", server.ShowAbout)
	r.GET("/artikel", server.ShowNews)
	r.GET("/kontak", server.ShowContact)
	r.GET("/faq", server.ShowFAQ)
	r.GET("/visi-misi", server.ShowVision)
	r.GET("/syarat-ketentuan", server.ShowTerms)
	r.GET("/privasi", server.ShowPrivacy)

	// Auth Routes
	r.GET("/login", server.ShowLogin)
	// Apply strict rate limiting to login POST (1 req/min, burst 5) to prevent brute force
	loginLimiter := middleware.RateLimiter(ctx, rate.Limit(1.0/60.0), 5)
	r.POST("/login", loginLimiter, server.Login)
	r.GET("/logout", server.Logout)

	// Apply rate limiting to captcha (10 req/min, burst 20) to prevent memory exhaustion DoS
	captchaLimiter := middleware.RateLimiter(ctx, rate.Limit(10.0/60.0), 20)
	r.GET("/api/captcha", captchaLimiter, server.GetCaptcha) // New endpoint for refreshing captcha

	// 6. Protected Routes
	protected := r.Group("/")
	protected.Use(middleware.AuthRequired(server.DB))
	{
		// Dashboard & Weighing
		// Note: The root "/" is now the public site. Dashboard is at /dashboard.
		protected.GET("/dashboard", server.ShowDashboard)
		protected.GET("/weighing", server.ShowWeighing)
		protected.GET("/reports", server.ShowReports)

		// API - Transactions & Hardware
		api := protected.Group("/api")
		{
			api.POST("/transaction", server.SaveTransaction)
			api.POST("/anpr/trigger", server.TriggerANPR)
			api.GET("/scales/stream", server.StreamScaleData)
			api.GET("/camera/stream", server.ProxyVideo)           // New RTSP proxy
			api.GET("/vehicles/details", server.GetVehicleDetails) // Allow operators to fetch details
			api.GET("/vehicles/search", server.SearchVehicles)     // Autocomplete
			api.GET("/reports/charts", server.GetReportCharts)     // Chart Data
		}

		// Admin Only Routes - Pages
		adminPages := protected.Group("/settings")
		adminPages.Use(middleware.RoleRequired("admin"))
		{
			adminPages.GET("/", server.ShowSettings)
			adminPages.GET("/vehicles", server.ShowVehicleSettings)
			adminPages.GET("/hardware", server.ShowSettingsHardware)
			adminPages.GET("/users", server.ShowUsers)
			adminPages.GET("/logs", server.ShowLogs)
		}

		// Admin Only Routes - APIs
		// We map them to /api/... but enforce admin role
		adminApi := protected.Group("/api")
		adminApi.Use(middleware.RoleRequired("admin"))
		{
			// Vehicle API
			adminApi.GET("/vehicles", server.ListVehicles)
			adminApi.POST("/vehicles", server.CreateVehicle)
			adminApi.PUT("/vehicles/:id", server.UpdateVehicle)
			adminApi.DELETE("/vehicles/:id", server.DeleteVehicle)

			// Company API
			adminApi.GET("/companies", server.ListCompanies)
			adminApi.POST("/companies", server.CreateCompany)
			adminApi.PUT("/companies/:id", server.UpdateCompany)
			adminApi.DELETE("/companies/:id", server.DeleteCompany)

			// Station / Hardware API
			adminApi.GET("/stations", server.GetStations)
			adminApi.POST("/stations", server.CreateStation)
			adminApi.PUT("/stations/:id", server.UpdateStation)
			adminApi.DELETE("/stations/:id", server.DeleteStation)

			// User Management API
			adminApi.GET("/users", server.GetUsers)
			adminApi.POST("/users", server.CreateUser)
			adminApi.DELETE("/users/:id", server.DeleteUser)
			adminApi.GET("/users/:id/assignments", server.GetUserAssignments)
			adminApi.POST("/users/:id/assignments", server.UpdateUserAssignments)

			// Logs
			adminApi.GET("/logs", server.GetLogsAPI)
		}
	}

	r.Any("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// External Device APIs (Token Based)
	r.POST("/api/external/scale", api.HandleRemoteScaleData)

	// 404 Handler
	r.NoRoute(server.Show404)

	return r
}
