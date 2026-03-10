package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/utils"
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

// setRefreshTokenCookie sets HttpOnly cookie for refresh token
func setRefreshTokenCookie(c *gin.Context, token string, maxAge int) {
	secure := os.Getenv("GO_ENV") == "production"
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
		secure, // secure flag - HTTPS only in production
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
		// Fallback: try to get from request body (backward compatibility)
		var req RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Refresh token not provided",
			})
			return
		}
		refreshToken = req.RefreshToken
	}

	// Get tenantID from context if available (set by middleware)
	tenantID := c.GetString("tenantID")

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
		_ = h.authService.Logout(c.Request.Context(), refreshToken)
	}

	// Clear the refresh token cookie
	clearRefreshTokenCookie(c)

	c.JSON(http.StatusOK, response.Success(gin.H{
		"message": "Logged out successfully",
	}))
}
