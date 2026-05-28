package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// UserSearchResult represents a user found during cross-tenant search.
type UserSearchResult struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	Status     string `json:"status"`
	TenantID   string `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
}

// SearchUsers searches users across active tenants by username or email.
// GET /api/dev/users/search?q=<query>&tenant_ids=<tenant_id>[,<tenant_id>]
func (h *DeveloperHandler) SearchUsers(c *gin.Context) {
	if !requireDeveloperPanelAccess(c) {
		return
	}
	query := strings.TrimSpace(c.Query("q"))
	if len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "query parameter 'q' must be at least 2 characters",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	explicitTenants := parseExplicitTenantScope(c.Query("tenant_ids"))
	if len(explicitTenants) == 0 {
		explicitTenants = parseExplicitTenantScope(c.GetHeader("x-tenant-scope"))
	}
	if len(explicitTenants) == 0 {
		// No explicit scope provided — default to all active tenants.
		// Each tenant is still gated by developerCanAccessTenant below.
		activeTenants, err := h.tenantService.GetAvailableTenants(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to list tenants: " + err.Error()})
			return
		}
		for _, t := range activeTenants {
			explicitTenants = append(explicitTenants, t.ID)
		}
	}
	if len(explicitTenants) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "no active tenants available"})
		return
	}
	for _, tenantID := range explicitTenants {
		if !developerCanAccessTenant(c, tenantID) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "tenant is outside authorized developer scope"})
			return
		}
	}

	var warning string
	if len(explicitTenants) > 50 {
		warning = "search limited to first 50 tenants"
		explicitTenants = explicitTenants[:50]
	}
	scopedTenants, err := h.tenantService.GetActiveTenantsByID(ctx, explicitTenants)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to list tenants: " + err.Error()})
		return
	}

	results := make([]UserSearchResult, 0)
	likePattern := "%" + query + "%"
	for _, tenantID := range explicitTenants {
		tenantInfo, tenantActive := scopedTenants[tenantID]
		if !tenantActive {
			continue
		}
		db, err := h.tenantService.GetTenantDB(tenantID)
		if err != nil {
			continue
		}
		var users []struct {
			ID       string `gorm:"column:id"`
			Username string `gorm:"column:username"`
			Email    string `gorm:"column:email"`
			Role     string `gorm:"column:role"`
		}
		db.WithContext(ctx).
			Table("users").
			Select("id, username, email, role").
			Where("username LIKE ? OR email LIKE ?", likePattern, likePattern).
			Limit(20).
			Find(&users)
		for _, u := range users {
			results = append(results, UserSearchResult{
				ID:         u.ID,
				Username:   u.Username,
				Email:      u.Email,
				Role:       u.Role,
				Status:     "active",
				TenantID:   tenantID,
				TenantName: tenantInfo.ShopName,
			})
		}
	}

	response := gin.H{"success": true, "data": results, "total": len(results)}
	if warning != "" {
		response["warning"] = warning
	}
	c.JSON(http.StatusOK, response)
}
