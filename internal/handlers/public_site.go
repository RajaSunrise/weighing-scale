package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// Public Site Handlers

func (s *Server) ShowHome(c *gin.Context) {
	c.HTML(http.StatusOK, "home.html", gin.H{})
}

func (s *Server) ShowProduct(c *gin.Context) {
	c.HTML(http.StatusOK, "produk.html", gin.H{})
}

func (s *Server) ShowGallery(c *gin.Context) {
	c.HTML(http.StatusOK, "galeri.html", gin.H{})
}

func (s *Server) ShowAbout(c *gin.Context) {
	c.HTML(http.StatusOK, "tentang.html", gin.H{})
}

func (s *Server) ShowNews(c *gin.Context) {
	c.HTML(http.StatusOK, "artikel.html", gin.H{})
}

func (s *Server) ShowContact(c *gin.Context) {
	c.HTML(http.StatusOK, "kontak.html", gin.H{})
}

func (s *Server) ShowFAQ(c *gin.Context) {
	c.HTML(http.StatusOK, "faq.html", gin.H{})
}

func (s *Server) ShowVision(c *gin.Context) {
	c.HTML(http.StatusOK, "visi-misi.html", gin.H{})
}

func (s *Server) ShowTerms(c *gin.Context) {
	c.HTML(http.StatusOK, "syarat-ketentuan.html", gin.H{})
}

func (s *Server) ShowPrivacy(c *gin.Context) {
	c.HTML(http.StatusOK, "privasi.html", gin.H{})
}
