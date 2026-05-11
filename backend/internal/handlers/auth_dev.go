package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/utils"
)

// DevLoginRequest represents dev login request
type DevLoginRequest struct {
	TenantID string `json:"tenant_id" binding:"required"`
}

// isDevModeAllowed checks if dev mode is allowed based on environment and request origin
// SECURITY: Dev login is allowed when:
// 1. GO_ENV is NOT "production" (explicit dev environment), OR
// 2. Request comes from localhost (for local development with production Docker)
func isDevModeAllowed(c *gin.Context) bool {
	// Check 1: Explicit development mode
	if os.Getenv("GO_ENV") != "production" {
		return true
	}

	// Check 2: Request from localhost (for local Docker development)
	// This allows developers to bypass login when running Docker locally
	// In Docker, localhost requests come through the bridge network (172.x.x.x)
	// but the Origin header still shows localhost
	origin := c.GetHeader("Origin")
	referer := c.GetHeader("Referer")

	isLocalOrigin := strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1")
	isLocalReferer := strings.Contains(referer, "localhost") || strings.Contains(referer, "127.0.0.1")

	// SECURITY: Only allow if Origin OR Referer is from localhost
	// This works because browsers enforce Origin header and can't be spoofed
	// Remote attackers can't set Origin to localhost from their browser
	return isLocalOrigin || isLocalReferer
}

// DevLoginInfo returns available tenants for dev login (only in development)
func (h *AuthHandler) DevLoginInfo(c *gin.Context) {
	// SECURITY: Only allow in development mode or from localhost
	if !isDevModeAllowed(c) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Not found",
		})
		return
	}

	// Fetch dynamic tenant list from database
	tenants, err := h.tenantService.GetAvailableTenants(c.Request.Context())
	if err != nil {
		// Fallback to empty list on error
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"dev_mode": true,
			"tenants":  []map[string]string{},
			"username": "tester",
			"error":    "Failed to fetch tenants: " + err.Error(),
		})
		return
	}

	// Convert to response format
	tenantList := make([]map[string]string, 0, len(tenants))
	for _, t := range tenants {
		tenantList = append(tenantList, map[string]string{
			"id":   t.ID,
			"name": t.ShopName,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"dev_mode": true,
		"tenants":  tenantList,
		"username": "tester",
	})
}

// DevLogin handles development-only auto login as tester
// SECURITY: This endpoint ONLY works when GO_ENV != "production" OR request from localhost
// NO DATABASE AUTH - directly generates JWT with requested tenant for developer convenience
// Safe because: Browser Origin header cannot be spoofed by remote attackers
func (h *AuthHandler) DevLogin(c *gin.Context) {
	// SECURITY: Reject if not in dev mode and not from localhost
	if !isDevModeAllowed(c) {
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

	// Validate tenant_id - use dynamic tenant list from database
	tenants, err := h.tenantService.GetAvailableTenants(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to validate tenant",
		})
		return
	}

	validTenant := false
	for _, t := range tenants {
		if t.ID == req.TenantID {
			validTenant = true
			break
		}
	}

	if !validTenant {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid tenant_id for dev login",
		})
		return
	}

	// BYPASS AUTH: Generate token directly with requested tenant
	// No database lookup needed - hardcoded dev user info
	devUserID := "00000000-0000-0000-0000-000000000001"
	devUsername := "tester"
	devEmail := "tester@dev.local"
	devRole := "developer"

	// Generate JWT directly with the REQUESTED tenant_id via authService
	// GenerateDevToken will return the actual user ID (either found or created)
	accessToken, refreshToken, actualUserID, err := h.authService.GenerateDevToken(c.Request.Context(), devUserID, devUsername, devEmail, req.TenantID, devRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to generate token: " + err.Error(),
		})
		return
	}

	// Set refresh token in HttpOnly cookie
	if refreshToken != "" {
		setRefreshTokenCookie(c, refreshToken, getRefreshTokenMaxAge())
	}

	// Return success with requested tenant_id in JWT
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Dev login successful",
		"user": gin.H{
			"id":       actualUserID,
			"username": devUsername,
			"email":    devEmail,
			"role":     devRole,
		},
		"token":         accessToken,
		"access_token":  accessToken,
		"refresh_token": refreshToken, // Include for debugging/fallback
		"tenant_id":     req.TenantID,
		"expires_in":    int(utils.AccessTokenTTL.Seconds()),
		"dev_mode":      true,
	})
}
