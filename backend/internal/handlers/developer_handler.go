package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services"
)

// DeveloperHandler provides developer-only cross-tenant operations.
// All routes are gated by RequireRole("developer") at the router level.
type DeveloperHandler struct {
	tenantService  *services.TenantService
	userMgmtService *services.UserManagementService
}

// NewDeveloperHandler creates a new developer handler.
func NewDeveloperHandler(tenantService *services.TenantService, userMgmtService *services.UserManagementService) *DeveloperHandler {
	return &DeveloperHandler{tenantService: tenantService, userMgmtService: userMgmtService}
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

// TenantDetailResponse represents a tenant with full details for the developer panel.
type TenantDetailResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	IsActive      bool    `json:"is_active"`
	UserCount     int64   `json:"user_count"`
	CreatedAt     string  `json:"created_at"`
	DeactivatedAt *string `json:"deactivated_at"`
}

// ListTenants returns all tenants (active and inactive) with user counts.
// GET /api/dev/tenants
func (h *DeveloperHandler) ListTenants(c *gin.Context) {
	ctx := c.Request.Context()

	systemDB, err := h.tenantService.GetSystemDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to get system db: " + err.Error(),
		})
		return
	}

	var tenants []struct {
		TenantID      string     `gorm:"column:tenant_id"`
		IsActive      bool       `gorm:"column:is_active"`
		CreatedAt     time.Time  `gorm:"column:created_at"`
		DeactivatedAt *time.Time `gorm:"column:deactivated_at"`
	}

	err = systemDB.WithContext(ctx).Table("system.tenants").
		Select("tenant_id, is_active, created_at, deactivated_at").
		Order("created_at ASC").
		Find(&tenants).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to list tenants: " + err.Error(),
		})
		return
	}

	results := make([]TenantDetailResponse, 0, len(tenants))
	for _, t := range tenants {
		detail := TenantDetailResponse{
			ID:        t.TenantID,
			Name:      t.TenantID,
			IsActive:  t.IsActive,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
		}

		if t.DeactivatedAt != nil {
			formatted := t.DeactivatedAt.Format(time.RFC3339)
			detail.DeactivatedAt = &formatted
		}

		// Count users in tenant schema
		db, err := h.tenantService.GetTenantDB(t.TenantID)
		if err == nil {
			repo := repositories.NewUserRepository(db)
			detail.UserCount, _ = repo.Count(ctx)
		}

		results = append(results, detail)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}

// resetPasswordRequest represents the request body for password reset.
type resetPasswordRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	TenantID    string `json:"tenant_id" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ResetPassword resets a user's password (developer cross-tenant action).
// POST /api/dev/reset-password
func (h *DeveloperHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation failed: " + err.Error(),
		})
		return
	}

	// Verify tenant exists by attempting to get its DB
	_, err := h.tenantService.GetTenantDB(req.TenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid tenant_id: " + err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resetBy := c.GetString("userID")
	err = h.userMgmtService.ResetPassword(ctx, req.UserID, req.NewPassword, resetBy, req.TenantID)
	if err != nil {
		switch err.Error() {
		case "user not found":
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "user not found",
			})
		default:
			// Password validation errors return 400
			if isPasswordValidationError(err) {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "failed to reset password: " + err.Error(),
				})
			}
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password reset successfully",
	})
}

// isPasswordValidationError checks if the error is from password strength validation.
func isPasswordValidationError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "password must") ||
		strings.Contains(msg, "password is too common")
}

// DeactivateTenant soft-deletes a tenant by setting is_active=false.
// DELETE /api/dev/tenants/:id
func (h *DeveloperHandler) DeactivateTenant(c *gin.Context) {
	tenantID := c.Param("id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "tenant id is required",
		})
		return
	}

	ctx := c.Request.Context()
	err := h.tenantService.DeactivateTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, services.ErrTenantNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "tenant not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to deactivate tenant: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// createTenantRequest represents the request body for tenant creation.
type createTenantRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateTenant creates a new tenant with schema and migrations.
// POST /api/dev/tenants
func (h *DeveloperHandler) CreateTenant(c *gin.Context) {
	var req createTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation failed: name is required",
		})
		return
	}

	ctx := c.Request.Context()
	result, err := h.tenantService.CreateTenant(ctx, req.Name)
	if err != nil {
		var validationErr *services.TenantValidationError
		var duplicateErr *services.TenantDuplicateError

		switch {
		case errors.As(err, &validationErr):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   validationErr.Error(),
			})
		case errors.As(err, &duplicateErr):
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   duplicateErr.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "failed to create tenant: " + err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    result,
	})
}

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

// SearchUsers searches users across all active tenants by username or email.
// GET /api/dev/users/search?q=<query>
func (h *DeveloperHandler) SearchUsers(c *gin.Context) {
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

	tenants, err := h.tenantService.GetAvailableTenants(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to list tenants: " + err.Error(),
		})
		return
	}

	var warning string
	if len(tenants) > 50 {
		warning = "search limited to first 50 tenants"
		tenants = tenants[:50]
	}

	results := make([]UserSearchResult, 0)
	likePattern := "%" + query + "%"

	for _, t := range tenants {
		db, err := h.tenantService.GetTenantDB(t.ID)
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
			status := "active"
			results = append(results, UserSearchResult{
				ID:         u.ID,
				Username:   u.Username,
				Email:      u.Email,
				Role:       u.Role,
				Status:     status,
				TenantID:   t.ID,
				TenantName: t.ShopName,
			})
		}
	}

	response := gin.H{
		"success": true,
		"data":    results,
		"total":   len(results),
	}
	if warning != "" {
		response["warning"] = warning
	}

	c.JSON(http.StatusOK, response)
}
