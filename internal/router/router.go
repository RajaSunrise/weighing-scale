package router

import (
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"stoneweigh/internal/handlers"
	"stoneweigh/internal/middleware"
)

func SetupRouter(server *handlers.Server) *gin.Engine {
	// 1. Initialize Gin
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 2. Setup Session Store
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "secret"
	}
	store := cookie.NewStore([]byte(secret))
	r.Use(sessions.Sessions("stoneweigh_session", store))

	// 3. Global Middleware
	r.Use(middleware.RequestLogger())
	// Rate Limit: 20 requests/second, burst of 50
	r.Use(middleware.RateLimiter(rate.Limit(20), 50))

	// 4. Static Files
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// 5. Public Routes
	r.GET("/login", server.ShowLogin)
	r.POST("/login", server.Login)
	r.GET("/logout", server.Logout)

	// 6. Protected Routes
	protected := r.Group("/")
	protected.Use(middleware.AuthRequired())
	{
		// Dashboard & Weighing
		protected.GET("/", server.ShowDashboard)
		protected.GET("/dashboard", server.ShowDashboard)
		protected.GET("/weighing", server.ShowWeighing)
		protected.GET("/reports", server.ShowReports)

		// API - Transactions & Hardware
		api := protected.Group("/api")
		{
			api.POST("/transaction", server.SaveTransaction)
			api.POST("/anpr/trigger", server.TriggerANPR)
			api.GET("/scales/stream", server.StreamScaleData)
		}

		// Admin Only Routes
		admin := protected.Group("/settings")
		admin.Use(middleware.RoleRequired("admin"))
		{
			admin.GET("/", server.ShowSettings)
			admin.GET("/vehicles", server.ShowVehicleSettings)
			admin.GET("/api/vehicles", server.ListVehicles)
			admin.POST("/api/vehicles", server.CreateVehicle)
			admin.DELETE("/api/vehicles/:id", server.DeleteVehicle)
		}
	}

	// 404 Handler
	r.NoRoute(server.Show404)

	return r
}
