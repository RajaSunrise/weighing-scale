package handlers

import (
	"fmt"
	"net/http"
	"time"

	"stoneweigh/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type ChartData struct {
	Labels []string  `json:"labels"`
	Data   []float64 `json:"data"`
}

// DailyStat holds the aggregation result from DB
type DailyStat struct {
	DateStr string  `gorm:"column:date_str"`
	Total   float64 `gorm:"column:total"`
}

func (s *Server) GetReportCharts(c *gin.Context) {
	period := c.Query("period") // daily, weekly, monthly

	now := time.Now()
	var startDate time.Time

	// Determine range
	switch period {
	case "weekly":
		// Last 12 weeks
		startDate = now.AddDate(0, 0, -84)
	case "monthly":
		// Last 12 months
		startDate = now.AddDate(-1, 0, 0)
	default: // daily
		// Last 30 days
		startDate = now.AddDate(0, 0, -30)
	}

	session := sessions.Default(c)

	// Optimization: Aggregation via SQL instead of fetching all records
	var stats []DailyStat

	// Determine Dialect for Date Function
	dateFunc := "TO_CHAR(weighed_at, 'YYYY-MM-DD')"
	if s.DB.Dialector.Name() == "sqlite" {
		dateFunc = "DATE(weighed_at)"
	}

	// SECURITY: Use GORM API with Scopes to apply station filtering (BOLA prevention)
	err := s.DB.Model(&models.WeighingRecord{}).
		Select(fmt.Sprintf("%s as date_str, SUM(net_weight) as total", dateFunc)).
		Where("weighed_at >= ?", startDate).
		Group("date_str").
		Scopes(s.stationFilter(session)).
		Scan(&stats).Error

	if err != nil {
		fmt.Printf("Error aggregating charts: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate charts"})
		return
	}

	// Map results: "YYYY-MM-DD" -> Total
	// Optimization: Pre-allocate map
	dayMap := make(map[string]float64, len(stats))
	for _, stat := range stats {
		dayMap[stat.DateStr] = stat.Total
	}

	// Optimization: Pre-allocate slices to avoid resizing
	// Max iterations is ~31 (Daily), so 32 is safe.
	labels := make([]string, 0, 32)
	data := make([]float64, 0, 32)

	if period == "daily" {
		cur := now.AddDate(0, 0, -30)
		for !cur.After(now) {
			dKey := cur.Format("2006-01-02")
			labels = append(labels, cur.Format("02 Jan"))
			data = append(data, dayMap[dKey])
			cur = cur.AddDate(0, 0, 1)
		}
	} else if period == "weekly" {
		// Aggregate by week
		// We start 12 weeks ago
		cur := now.AddDate(0, 0, -84)
		// Align to Monday?
		for cur.Weekday() != time.Monday {
			cur = cur.AddDate(0, 0, -1)
		}

		for !cur.After(now) {
			weekSum := 0.0
			weekLabel := cur.Format("02 Jan")
			// Sum next 7 days
			for i := range 7 {
				dKey := cur.AddDate(0, 0, i).Format("2006-01-02")
				weekSum += dayMap[dKey]
			}
			labels = append(labels, weekLabel)
			data = append(data, weekSum)
			cur = cur.AddDate(0, 0, 7)
		}
	} else if period == "monthly" {
		// Aggregate by month
		cur := now.AddDate(-1, 0, 0)
		// Align to 1st
		cur = time.Date(cur.Year(), cur.Month(), 1, 0, 0, 0, 0, cur.Location())

		for !cur.After(now) {
			monSum := 0.0
			monLabel := cur.Format("Jan '06")
			// Sum all days in this month
			nextMonth := cur.AddDate(0, 1, 0)
			temp := cur
			for temp.Before(nextMonth) {
				dKey := temp.Format("2006-01-02")
				monSum += dayMap[dKey]
				temp = temp.AddDate(0, 0, 1)
			}
			labels = append(labels, monLabel)
			data = append(data, monSum)
			cur = nextMonth
		}
	}

	c.JSON(http.StatusOK, ChartData{
		Labels: labels,
		Data:   data,
	})
}
