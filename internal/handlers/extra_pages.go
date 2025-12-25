package handlers

import (
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"stoneweigh/internal/models"
)

// ShowReports renders the reports page with historical data
func (s *Server) ShowReports(c *gin.Context) {
	session := sessions.Default(c)
	fullName := "Operator"
	if v := session.Get("full_name"); v != nil {
		fullName = v.(string)
	}

	// Parse filters
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			start = time.Now().AddDate(0, 0, -30) // Default 30 days
		}
	} else {
		start = time.Now().AddDate(0, 0, -7) // Default 7 days
	}

	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			end = time.Now()
		}
	} else {
		end = time.Now()
	}

	// Adjust end to end of day
	end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	var records []models.WeighingRecord
	s.DB.Where("weighed_at BETWEEN ? AND ?", start, end).Order("weighed_at desc").Find(&records)

	c.HTML(http.StatusOK, "reports.html", gin.H{
		"title":       "Laporan",
		"active":      "reports",
		"showNav":     true,
		"CurrentUser": fullName,
		"Records":     records,
		"StartDate":   start.Format("2006-01-02"),
		"EndDate":     end.Format("2006-01-02"),
	})
}

// ShowSettings renders the main settings landing page
func (s *Server) ShowSettings(c *gin.Context) {
	session := sessions.Default(c)
	fullName := "Operator"
	if v := session.Get("full_name"); v != nil {
		fullName = v.(string)
	}
	c.HTML(http.StatusOK, "settings.html", gin.H{
		"title":       "Settings",
		"active":      "settings",
		"showNav":     true,
		"CurrentUser": fullName,
	})
}

// Show404 renders the custom not found page
func (s *Server) Show404(c *gin.Context) {
	c.HTML(http.StatusNotFound, "404.html", gin.H{
		"title": "Halaman Tidak Ditemukan",
	})
}
