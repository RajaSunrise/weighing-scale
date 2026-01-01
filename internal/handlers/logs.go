package handlers

import (
	"bufio"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// ShowLogs renders the log viewer page
func (s *Server) ShowLogs(c *gin.Context) {
	session := sessions.Default(c)
	fullName := "Operator"
	if v := session.Get("full_name"); v != nil {
		fullName = v.(string)
	}

	c.HTML(http.StatusOK, "logs.html", gin.H{
		"title":       "System Logs",
		"active":      "settings",
		"showNav":     true,
		"CurrentUser": fullName,
	})
}

// GetLogsAPI returns the last N lines of the log file
func (s *Server) GetLogsAPI(c *gin.Context) {
	file, err := os.Open("logs/system.log")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open log file"})
		return
	}
	defer file.Close()

	// Simple implementation: Read all and return last 100 lines
	// For large logs, seek from end is better, but this is sufficient for MVP.
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	start := max(len(lines)-100, 0)

	c.JSON(http.StatusOK, lines[start:])
}
