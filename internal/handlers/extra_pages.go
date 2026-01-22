package handlers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

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
