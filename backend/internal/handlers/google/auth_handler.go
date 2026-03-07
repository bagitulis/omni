// Package google handles Google API related endpoints
package google

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/google"
)

// AuthHandler handles Google authentication endpoints
type AuthHandler struct {
	authService *google.AuthService
}

// NewAuthHandler creates a new Google auth handler
func NewAuthHandler(authService *google.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// GetAuthStatus handles GET /api/google/auth/status
// Returns the current Google Sheets service account status
func (h *AuthHandler) GetAuthStatus(c *gin.Context) {
	status := h.authService.GetStatus()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"authenticated":   status.IsAuthenticated,
			"service_account": status.ServiceAccount,
			"email":           status.Email,
			"scopes":          status.Scopes,
			"expires_at":      status.ExpiresAt,
		},
	})
}
