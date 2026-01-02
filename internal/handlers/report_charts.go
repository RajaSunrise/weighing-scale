package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"stoneweigh/internal/models"
)

type ChartData struct {
	Labels []string  `json:"labels"`
	Data   []float64 `json:"data"`
}

func (s *Server) GetReportCharts(c *gin.Context) {
	period := c.Query("period") // daily, weekly, monthly

	now := time.Now()
	var startDate time.Time

	// Determine range and grouping
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

	var records []models.WeighingRecord
	s.DB.Select("weighed_at, net_weight").Where("weighed_at >= ?", startDate).Find(&records)

	// Re-implementation of aggregation logic to be robust
	// 1. Create a map of "YYYY-MM-DD" -> NetWeight
	dayMap := make(map[string]float64)
	for _, r := range records {
		dayMap[r.WeighedAt.Format("2006-01-02")] += r.NetWeight
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
		cur = time.Date(cur.Year(), cur.Month(), 1, 0,0,0,0, cur.Location())

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
