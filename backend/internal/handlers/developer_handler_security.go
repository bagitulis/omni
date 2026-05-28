package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
)

const (
	developerScopeContextKey        = "developer_tenant_scope"
	developerImpersonatedContextKey = "impersonated"
	roleSuperadmin                  = "superadmin"
	developerAuditTenant            = "system"
	tenantScopeAll                  = "*"
)

func requireDeveloperPanelAccess(c *gin.Context) bool {
	if isImpersonatedRequest(c) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "developer endpoints are blocked during impersonation"})
		return false
	}
	role := c.GetString("role")
	if role != models.RoleDeveloper && role != models.RoleService && role != roleSuperadmin {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "developer role required"})
		return false
	}
	return true
}

func isImpersonatedRequest(c *gin.Context) bool {
	for _, key := range []string{developerImpersonatedContextKey, "is_impersonated", "impersonation_active"} {
		if value, exists := c.Get(key); exists {
			if flag, ok := value.(bool); ok && flag {
				return true
			}
		}
	}
	return c.GetString("impersonated_by") != "" || c.GetString("impersonation_actor_id") != ""
}

func developerAllowedTenants(c *gin.Context) map[string]struct{} {
	allowed := map[string]struct{}{}
	add := func(tenantID string) {
		if trimmed := strings.TrimSpace(tenantID); trimmed != "" {
			allowed[trimmed] = struct{}{}
		}
	}
	if value, exists := c.Get(developerScopeContextKey); exists {
		switch tenants := value.(type) {
		case []string:
			for _, tenantID := range tenants {
				add(tenantID)
			}
		case string:
			for _, tenantID := range strings.Split(tenants, ",") {
				add(tenantID)
			}
		}
	}
	if len(allowed) == 0 {
		add(c.GetString("tenant_id"))
	}
	return allowed
}

func developerHasExplicitTenantScope(c *gin.Context) bool {
	if c.GetString("role") == models.RoleService || c.GetString("role") == roleSuperadmin {
		return true
	}
	allowed := developerAllowedTenants(c)
	_, hasAll := allowed[tenantScopeAll]
	return len(allowed) > 0 && !hasAll
}

func developerCanAccessTenant(c *gin.Context, tenantID string) bool {
	role := c.GetString("role")
	if role == models.RoleService || role == roleSuperadmin {
		return true
	}
	allowed := developerAllowedTenants(c)
	if _, all := allowed[tenantScopeAll]; all {
		return true
	}
	_, ok := allowed[tenantID]
	return ok
}

func requireDeveloperMutationScope(c *gin.Context, tenantID string) bool {
	if !requireDeveloperTenantAccess(c, tenantID) {
		return false
	}
	if !developerHasExplicitTenantScope(c) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "explicit developer tenant scope is required for tenant mutation"})
		return false
	}
	return true
}

func requireDeveloperTenantAccess(c *gin.Context, tenantID string) bool {
	if !requireDeveloperPanelAccess(c) {
		return false
	}
	if !developerCanAccessTenant(c, tenantID) {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "tenant is outside authorized developer scope"})
		return false
	}
	return true
}

func canDeveloperGrantRole(actorRole, actorID, targetUserID, targetRole string) bool {
	if targetRole == "" {
		return true
	}
	if targetRole == roleSuperadmin || targetRole == models.RoleDeveloper || targetRole == models.RoleService {
		return actorRole == roleSuperadmin || actorRole == models.RoleService
	}
	if actorRole == models.RoleDeveloper && actorID == targetUserID && targetRole == models.RoleAdmin {
		return false
	}
	return actorRole == models.RoleDeveloper || actorRole == roleSuperadmin || actorRole == models.RoleService
}

func parseExplicitTenantScope(raw string) []string {
	seen := map[string]struct{}{}
	var tenants []string
	for _, part := range strings.Split(raw, ",") {
		tenantID := strings.TrimSpace(part)
		if tenantID == "" {
			continue
		}
		if _, exists := seen[tenantID]; exists {
			continue
		}
		seen[tenantID] = struct{}{}
		tenants = append(tenants, tenantID)
	}
	return tenants
}

func (h *DeveloperHandler) writeDeveloperAudit(c *gin.Context, action, status, targetUserID, targetTenantID string, details map[string]interface{}, errMsg string) {
	systemDB, err := h.tenantService.GetSystemDB()
	if err != nil {
		return
	}
	if details == nil {
		details = map[string]interface{}{}
	}
	details["actor_role"] = c.GetString("role")
	repo := repositories.NewAuditRepository(systemDB)
	_ = repo.Create(c.Request.Context(), &models.AuditLogEntry{
		TenantID:       developerAuditTenant,
		Action:         action,
		UserID:         c.GetString("userID"),
		TargetUserID:   targetUserID,
		TargetTenantID: targetTenantID,
		Details:        details,
		Status:         status,
		ErrorMessage:   errMsg,
		IPAddress:      c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
	})
}

func auditDetails(values map[string]interface{}) map[string]interface{} {
	if values == nil {
		return map[string]interface{}{}
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return map[string]interface{}{"metadata": "unavailable"}
	}
	var sanitized map[string]interface{}
	if err := json.Unmarshal(encoded, &sanitized); err != nil {
		return map[string]interface{}{"metadata": "unavailable"}
	}
	return sanitized
}
