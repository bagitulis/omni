package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/utils"
)

// DevLoginRequest represents dev login request
type DevLoginRequest struct {
	TenantID string `json:"tenant_id" binding:"required"`
}

// DevLoginInfo returns available tenants for dev login (only in development)
func (h *AuthHandler) DevLoginInfo(c *gin.Context) {
	// SECURITY: Only allow in development mode
	if os.Getenv("GO_ENV") == "production" {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Not found",
		})
		return
	}

	// Return available tenants for dev mode
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"dev_mode": true,
		"tenants": []map[string]string{
			{"id": "yumna_bertigamart", "name": "Yumna - Bertigamart"},
			{"id": "tika_nusseyba", "name": "Tika - Nusseyba"},
		},
		"username": "tester",
	})
}

// DevLogin handles development-only auto login as tester
// SECURITY: This endpoint ONLY works when GO_ENV != "production"
func (h *AuthHandler) DevLogin(c *gin.Context) {
	// SECURITY: Reject in production mode - this is the critical check
	if os.Getenv("GO_ENV") == "production" {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Not found",
		})
		return
	}

	var req DevLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "tenant_id is required",
		})
		return
	}

	// Validate tenant_id - only allow known test tenants
	validTenants := map[string]bool{
		"yumna_bertigamart": true,
		"tika_nusseyba":     true,
	}

	if !validTenants[req.TenantID] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid tenant_id for dev login",
		})
		return
	}

	// Use internal login with hardcoded tester credentials
	// These credentials are only used server-side, never exposed to frontend
	result, err := h.multiTenantAuth.LoginAcrossTenants(c.Request.Context(), &services.MultiTenantLoginRequest{
		Username:  "tester",
		Password:  "tester@123",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent() + " (DevLogin)",
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Dev login failed: " + err.Error(),
		})
		return
	}

	// Set refresh token cookie
	if result.RefreshToken != "" {
		setRefreshTokenCookie(c, result.RefreshToken, getRefreshTokenMaxAge())
	}

	// Return success with requested tenant_id
	// Note: tester account exists in both tenants, so we use the requested tenant
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"message":      "Dev login successful",
		"user":         result.User,
		"token":        result.AccessToken,
		"access_token": result.AccessToken,
		"tenant_id":    req.TenantID,
		"expires_in":   int(utils.AccessTokenTTL.Seconds()),
		"dev_mode":     true,
	})
}
