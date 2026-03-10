package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/services"
)

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

	c.JSON(http.StatusOK, response.Success(gin.H{
		"tenants": result,
	}))
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

	c.JSON(http.StatusOK, response.Success(gin.H{
		"message":   "Switched to tenant '" + req.TenantID + "'",
		"token":     newToken,
		"tenant_id": req.TenantID,
	}))
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
