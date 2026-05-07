package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
)

// CSRFHandler handles CSRF token endpoints
type CSRFHandler struct{}

// NewCSRFHandler creates a new CSRF handler
func NewCSRFHandler() *CSRFHandler {
	return &CSRFHandler{}
}

// GetCSRFToken generates and returns a new CSRF token
// GET /api/csrf-token
func (h *CSRFHandler) GetCSRFToken(c *gin.Context) {
	token, err := middleware.GenerateCSRFToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate CSRF token",
		})
		return
	}

	// Set cookie with the token (for double-submit pattern)
	// httpOnly MUST be false so frontend JavaScript can read it
	// and include it in the x-csrf-token header
	c.SetCookie(
		"csrf_token",
		token,
		int(24*time.Hour/time.Second), // 24 hours
		"/",
		"",    // domain (empty = current domain)
		false, // secure (should be true in production with HTTPS)
		false, // httpOnly = false (JS must read this for double-submit pattern)
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token":      token,
			"expires_in": 86400, // 24 hours in seconds
		},
	})
}
