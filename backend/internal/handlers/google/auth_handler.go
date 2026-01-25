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

// GetAuthURL handles GET /api/google/auth/url
// Returns OAuth URL for Google authentication (if using OAuth flow)
func (h *AuthHandler) GetAuthURL(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId - authentication required",
		})
		return
	}

	redirectURI := c.Query("redirect_uri")
	if redirectURI == "" {
		redirectURI = c.Request.Host + "/api/google/auth/callback"
	}

	url, state, err := h.authService.GenerateOAuthURL(redirectURI, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"url":   url,
			"state": state,
		},
	})
}

// HandleCallback handles GET /api/google/auth/callback
// Processes OAuth callback from Google
func (h *AuthHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	if errorParam != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "OAuth error: " + errorParam,
		})
		return
	}

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing authorization code",
		})
		return
	}

	// Validate state and exchange code for token
	tenantID, err := h.authService.ValidateOAuthState(state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid state: " + err.Error(),
		})
		return
	}

	token, err := h.authService.ExchangeCode(c.Request.Context(), code, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to exchange code: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Google authentication successful",
		"data": gin.H{
			"tenant_id":  tenantID,
			"expires_at": token.Expiry,
		},
	})
}

// Disconnect handles POST /api/google/auth/disconnect
// Disconnects Google account for tenant
func (h *AuthHandler) Disconnect(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId - authentication required",
		})
		return
	}

	if err := h.authService.RevokeAccess(tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Google account disconnected",
	})
}
