package handlers

import (
	"bytes"
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

	stat, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stat log file"})
		return
	}
	filesize := stat.Size()

	// If file is empty
	if filesize == 0 {
		c.JSON(http.StatusOK, []string{})
		return
	}

	var lines []string
	const bufferSize = 4096
	buffer := make([]byte, bufferSize)

	cursor := filesize
	var leftover []byte

	// Read backwards
	for cursor > 0 && len(lines) < 100 {
		toRead := bufferSize
		if int64(toRead) > cursor {
			toRead = int(cursor)
		}
		cursor -= int64(toRead)

		_, err := file.Seek(cursor, 0)
		if err != nil {
			break
		}

		n, err := file.Read(buffer[:toRead])
		if err != nil {
			break
		}

		chunk := buffer[:n]

		// If this is the end of the file and it ends with \n, ignore the last \n
		// because it's just a line terminator, not a separator for a new empty line
		if cursor+int64(n) == filesize && n > 0 && chunk[n-1] == '\n' {
			chunk = chunk[:n-1]
		}

		// Combine with leftover from previous iteration (which was "after" this chunk)
		data := append(chunk, leftover...)

		parts := bytes.Split(data, []byte{'\n'})

		// The first part is the new leftover (start of the chunk/line)
		leftover = parts[0]

		// The rest are complete lines (in forward order)
		// We process them in reverse to add to our result
		for i := len(parts) - 1; i > 0; i-- {
			lines = append([]string{string(parts[i])}, lines...)
			if len(lines) >= 100 {
				break
			}
		}
	}

	// Add the final leftover as the first line if we still need lines
	if len(lines) < 100 {
		lines = append([]string{string(leftover)}, lines...)
	}

	c.JSON(http.StatusOK, lines)
}
