package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/utils"
	"github.com/rs/zerolog/log"
)

// Cookie constants for refresh token
const (
	RefreshTokenCookieName = "refresh_token"
	RefreshTokenCookiePath = "/api/auth"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService           *services.AuthService
	multiTenantAuth       *services.MultiTenantAuthService
	userManagementService *services.UserManagementService
	basePath              string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService, multiTenantAuth *services.MultiTenantAuthService, userManagementService *services.UserManagementService, basePath string) *AuthHandler {
	return &AuthHandler{
		authService:           authService,
		multiTenantAuth:       multiTenantAuth,
		userManagementService: userManagementService,
		basePath:              basePath,
	}
}

// isLocalhostRequest checks if the request originates from localhost.
// Used to allow HTTP cookies for local Docker development.
func isLocalhostRequest(c *gin.Context) bool {
	origin := c.GetHeader("Origin")
	referer := c.GetHeader("Referer")
	return strings.Contains(origin, "localhost") ||
		strings.Contains(origin, "127.0.0.1") ||
		strings.Contains(referer, "localhost") ||
		strings.Contains(referer, "127.0.0.1")
}

// setRefreshTokenCookie sets HttpOnly cookie for refresh token.
// HYBRID MODE: When GO_ENV=production but request is from localhost (Docker dev),
// Secure=false so the browser accepts the cookie over HTTP.
func setRefreshTokenCookie(c *gin.Context, token string, maxAge int) {
	isProd := os.Getenv("GO_ENV") == "production"
	isLocal := isLocalhostRequest(c)

	// Secure=true only for production HTTPS, not for localhost HTTP
	secure := isProd && !isLocal
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteStrictMode
	}

	c.SetSameSite(sameSite)
	c.SetCookie(
		RefreshTokenCookieName,
		token,
		maxAge,
		RefreshTokenCookiePath,
		"",     // domain - empty uses current domain
		secure, // secure flag - HTTPS only in non-localhost production
		true,   // httpOnly - CRITICAL for security!
	)
}

// clearRefreshTokenCookie clears the refresh token cookie
func clearRefreshTokenCookie(c *gin.Context) {
	c.SetCookie(RefreshTokenCookieName, "", -1, RefreshTokenCookiePath, "", false, true)
}

// getRefreshTokenMaxAge returns max age in seconds for refresh token cookie
func getRefreshTokenMaxAge() int {
	return int(utils.RefreshTokenTTL.Seconds())
}

// Login handles user login - searches across ALL tenants (matches Node.js findUserAcrossTenants)
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Username and password are required",
		})
		return
	}

	// Use multi-tenant login (searches all tenants)
	result, err := h.multiTenantAuth.LoginAcrossTenants(c.Request.Context(), &services.MultiTenantLoginRequest{
		Username:     req.Username,
		Password:     req.Password,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		CaptchaToken: req.RecaptchaToken,
	})

	if err != nil {
		statusCode := http.StatusUnauthorized
		if authErr, ok := err.(*services.AuthError); ok {
			if authErr.Code == "ACCOUNT_LOCKED" {
				statusCode = http.StatusForbidden
			}
		}
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Set refresh token in HttpOnly cookie if available
	if result.RefreshToken != "" {
		setRefreshTokenCookie(c, result.RefreshToken, getRefreshTokenMaxAge())
	}

	// Response format matches Node.js but uses snake_case per AGENTS.MD
	// Access token is returned in body, refresh token is in HttpOnly cookie
	c.JSON(http.StatusOK, response.Success(gin.H{
		"message":      "Login successful",
		"user":         result.User,
		"token":        result.AccessToken,
		"access_token": result.AccessToken,
		"tenant_id":    result.TenantID,
		"expires_in":   int(utils.AccessTokenTTL.Seconds()),
	}))
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// First try to get refresh token from HttpOnly cookie (secure method)
	refreshToken, err := c.Cookie(RefreshTokenCookieName)
	if err != nil || refreshToken == "" {
		// Fallback: try to get from request body (DEPRECATED — will be removed)
		var req RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Refresh token not provided",
			})
			return
		}
		refreshToken = req.RefreshToken
		// Signal deprecation to callers
		c.Header("Deprecation", "true")
		c.Header("Sunset", "2026-06-01")
		log.Warn().Msg("Refresh token received via request body (deprecated) — migrate to HttpOnly cookie")
	}

	// Get tenantID from context if available (set by middleware)
	tenantID := middleware.GetTenantID(c)

	// Use multi-tenant refresh with rotation (tries tenant DB first, then system DB)
	accessToken, newRefreshToken, err := h.multiTenantAuth.RefreshTokenForTenant(
		c.Request.Context(),
		refreshToken,
		tenantID,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	if err != nil {
		// Check for token reuse attack
		if authErr, ok := err.(*services.AuthError); ok && authErr.Code == "TOKEN_REUSE" {
			// Clear the compromised cookie
			clearRefreshTokenCookie(c)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Session compromised. Please login again.",
				"code":    "TOKEN_REUSE",
			})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid or expired refresh token",
		})
		return
	}

	// Set new refresh token cookie if rotation occurred
	if newRefreshToken != "" {
		setRefreshTokenCookie(c, newRefreshToken, getRefreshTokenMaxAge())
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"access_token": accessToken,
		"expires_in":   int(utils.AccessTokenTTL.Seconds()),
	}))
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get refresh token from cookie
	refreshToken, _ := c.Cookie(RefreshTokenCookieName)

	// Revoke the refresh session if token exists
	if refreshToken != "" {
		if err := h.authService.Logout(c.Request.Context(), refreshToken); err != nil {
			// Log the failure for security auditing — token may still be valid
			log.Warn().Err(err).Msg("Failed to revoke refresh token during logout")
		}
	}

	// Always clear the cookie regardless of revocation result
	clearRefreshTokenCookie(c)

	c.JSON(http.StatusOK, response.Success(gin.H{
		"message": "Logged out successfully",
	}))
}
