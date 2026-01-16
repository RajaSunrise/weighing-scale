package handlers

import (
	"fmt"
	"net/http"
	"time"

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

	// Optimization: Aggregation via SQL instead of fetching all records
	var stats []DailyStat
	var err error

	// Determine Dialect for Date Function
	if s.DB.Dialector.Name() == "sqlite" {
		// SQLite: DATE(weighed_at) returns string "YYYY-MM-DD"
		err = s.DB.Raw("SELECT DATE(weighed_at) as date_str, SUM(net_weight) as total FROM weighing_records WHERE weighed_at >= ? GROUP BY 1", startDate).Scan(&stats).Error
	} else {
		// Postgres: TO_CHAR(weighed_at, 'YYYY-MM-DD') returns string
		err = s.DB.Raw("SELECT TO_CHAR(weighed_at, 'YYYY-MM-DD') as date_str, SUM(net_weight) as total FROM weighing_records WHERE weighed_at >= ? GROUP BY 1", startDate).Scan(&stats).Error
	}

	if err != nil {
		fmt.Printf("Error aggregating charts: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate charts"})
		return
	}

	// Map results: "YYYY-MM-DD" -> Total
	dayMap := make(map[string]float64)
	for _, stat := range stats {
		dayMap[stat.DateStr] = stat.Total
	}

	labels := []string{}
	data := []float64{}

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
			for i := 0; i < 7; i++ {
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
