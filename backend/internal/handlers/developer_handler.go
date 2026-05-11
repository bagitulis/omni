package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services"
)

// DeveloperHandler provides developer-only cross-tenant operations.
// All routes are gated by RequireRole("developer") at the router level.
type DeveloperHandler struct {
	tenantService *services.TenantService
}

// NewDeveloperHandler creates a new developer handler.
func NewDeveloperHandler(tenantService *services.TenantService) *DeveloperHandler {
	return &DeveloperHandler{tenantService: tenantService}
}

// TenantOverview is a per-tenant summary for the developer panel.
type TenantOverview struct {
	ID        string `json:"id"`
	ShopName  string `json:"shop_name"`
	UserCount int64  `json:"user_count"`
	Owners    int64  `json:"owners"`
	Admins    int64  `json:"admins"`
	Users     int64  `json:"users"`
	Error     string `json:"error,omitempty"`
}

// GetOverview returns list of tenants with user counts grouped by role.
// GET /api/dev/overview
func (h *DeveloperHandler) GetOverview(c *gin.Context) {
	ctx := c.Request.Context()

	tenants, err := h.tenantService.GetAvailableTenants(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to list tenants: " + err.Error(),
		})
		return
	}

	overviews := make([]TenantOverview, 0, len(tenants))
	for _, t := range tenants {
		overview := TenantOverview{
			ID:       t.ID,
			ShopName: t.ShopName,
		}

		db, err := h.tenantService.GetTenantDB(t.ID)
		if err != nil {
			overview.Error = "failed to open tenant db: " + err.Error()
			overviews = append(overviews, overview)
			continue
		}

		repo := repositories.NewUserRepository(db)
		overview.UserCount, _ = repo.Count(ctx)
		overview.Owners, _ = repo.CountByRole(ctx, "owner")
		overview.Admins, _ = repo.CountByRole(ctx, "admin")
		overview.Users, _ = repo.CountByRole(ctx, "user")

		overviews = append(overviews, overview)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tenants": overviews,
			"total":   len(overviews),
		},
	})
}
