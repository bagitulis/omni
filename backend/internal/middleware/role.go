package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
)

// RoleHierarchy defines the privilege level for each role.
// Higher number = more privilege.
var RoleHierarchy = map[string]int{
	models.RoleUser:      1,
	models.RoleAdmin:     2,
	models.RoleOwner:     3,
	models.RoleDeveloper: 4,
	models.RoleService:   5,
}

// RequireRole returns middleware that checks if the authenticated user
// has one of the allowed roles. Must be used AFTER Auth() middleware.
func RequireRole(allowed ...string) gin.HandlerFunc {
	allowedSet := make(map[string]bool, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = true
	}

	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied: no role in context",
			})
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied: invalid role format",
			})
			c.Abort()
			return
		}

		if !allowedSet[roleStr] {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Access denied: insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CanManageRole checks if the acting user's role can manage the target role.
// Rules:
//   - developer: can manage all roles (cross-tenant super admin)
//   - owner: can manage admin, user
//   - admin: can manage user only
//   - user/service: cannot manage anyone
func CanManageRole(actorRole, targetRole string) bool {
	actorLevel := RoleHierarchy[actorRole]
	targetLevel := RoleHierarchy[targetRole]

	if actorLevel == 0 || targetLevel == 0 {
		return false
	}

	// Actor must have strictly higher privilege than target
	return actorLevel > targetLevel
}
