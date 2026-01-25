package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services"
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

// Login handles user login - searches across ALL tenants
// Matches Node.js behavior: findUserAcrossTenants
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

	// Response format matches Node.js but uses snake_case per AGENTS.MD
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Login successful",
		"user":      result.User,
		"token":     result.AccessToken,
		"tenant_id": result.TenantID,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	accessToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token": accessToken,
		},
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT setup, logout is typically handled client-side
	// by removing the token. Server-side logout would require token blacklisting.
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User not authenticated",
		})
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password changed successfully",
	})
}

// GetCurrentUser returns current authenticated user info
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID := c.GetString("userID")
	tenantID := c.GetString("tenantID")
	role := c.GetString("role")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user_id":   userID,
			"tenant_id": tenantID,
			"role":      role,
		},
	})
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if h.userManagementService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "User management service not configured",
		})
		return
	}

	user, err := h.userManagementService.CreateUser(c.Request.Context(), &services.CreateUserRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     "owner",
	}, "system")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User registered successfully",
		"user":    user.ToResponse(),
	})
}

// GetTenants returns available tenants for switching
func (h *AuthHandler) GetTenants(c *gin.Context) {
	tenants, err := h.multiTenantAuth.GetAvailableTenants(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to load tenants: " + err.Error(),
		})
		return
	}

	// Convert to response format (snake_case per AGENTS.MD)
	var result []gin.H
	for _, t := range tenants {
		result = append(result, gin.H{
			"id":        t.ID,
			"shop_name": t.ShopName,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"tenants": result,
	})
}

// SwitchTenant switches to a different tenant (developer only)
func (h *AuthHandler) SwitchTenant(c *gin.Context) {
	var req SwitchTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "tenant_id is required",
		})
		return
	}

	role := c.GetString("role")
	userID := c.GetString("userID")

	// Use multiTenantAuth for switch
	newToken, err := h.multiTenantAuth.SwitchTenant(c.Request.Context(), userID, role, req.TenantID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if authErr, ok := err.(*services.AuthError); ok {
			switch authErr.Code {
			case "FORBIDDEN":
				statusCode = http.StatusForbidden
			case "TENANT_NOT_FOUND":
				statusCode = http.StatusBadRequest
			}
		}
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Switched to tenant '" + req.TenantID + "'",
		"token":     newToken,
		"tenant_id": req.TenantID,
	})
}

// VerifyToken verifies token validity
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"valid":   false,
			"error":   "No token provided",
		})
		return
	}

	// Extract token from Bearer header
	token := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	}

	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"valid":   false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"valid":   true,
		"payload": claims,
	})
}
