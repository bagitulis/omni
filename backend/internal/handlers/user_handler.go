package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services"
)

// UserHandler handles user management endpoints
type UserHandler struct {
	userService   *services.UserManagementService // fallback service (system scope, used only as safety net)
	tenantService *services.TenantService
	auditRepo     *repositories.AuditRepository
	authService   *services.AuthService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserManagementService, tenantService *services.TenantService, auditRepo *repositories.AuditRepository, authService *services.AuthService) *UserHandler {
	return &UserHandler{
		userService:   userService,
		tenantService: tenantService,
		auditRepo:     auditRepo,
		authService:   authService,
	}
}

// getTenantService builds a tenant-scoped UserManagementService for the current request.
// Uses the tenant DB for user queries so that users are properly scoped to the tenant schema.
func (h *UserHandler) getTenantService(tenantID string) (*services.UserManagementService, error) {
	tenantDB, err := h.tenantService.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}
	userRepo := repositories.NewUserRepository(tenantDB)
	return services.NewUserManagementService(userRepo, h.auditRepo, h.authService), nil
}

// CreateUserRequest represents create user request body
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required"`
}

// CreateUser creates a new user
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate target role
	if !models.IsValidRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid role: " + req.Role,
		})
		return
	}

	// Enforce role hierarchy: actor can only create users with lower privilege
	actorRole := c.GetString("role")
	if !middleware.CanManageRole(actorRole, req.Role) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Cannot create user with role '" + req.Role + "': insufficient permissions",
		})
		return
	}

	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}

	createdBy := c.GetString("userID")

	userService, err := h.getTenantService(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to access tenant database: " + err.Error(),
		})
		return
	}

	user, err := userService.CreateUser(c.Request.Context(), &services.CreateUserRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
		TenantID: tenantID,
	}, createdBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    user.ToResponse(),
	})
}

// UpdateUserRequest represents update user request body
type UpdateUserRequest struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
}

// UpdateUser updates an existing user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "User ID is required",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// If role is being changed, enforce hierarchy
	actorRole := c.GetString("role")
	if req.Role != "" {
		if !models.IsValidRole(req.Role) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid role: " + req.Role,
			})
			return
		}
		if !middleware.CanManageRole(actorRole, req.Role) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Cannot assign role '" + req.Role + "': insufficient permissions",
			})
			return
		}
	}

	// Get tenant-scoped service for all user operations
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}
	userService, err := h.getTenantService(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to access tenant database: " + err.Error(),
		})
		return
	}

	// Also check: actor must be able to manage the user's CURRENT role
	// (e.g., admin cannot edit an owner's profile)
	existingUser, err := userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if existingUser == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}
	if !middleware.CanManageRole(actorRole, existingUser.Role) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Cannot modify user with role '" + existingUser.Role + "': insufficient permissions",
		})
		return
	}

	updatedBy := c.GetString("userID")

	user, err := userService.UpdateUser(c.Request.Context(), userID, &services.UpdateUserRequest{
		Username: req.Username,
		Email:    req.Email,
		Role:     req.Role,
	}, updatedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user.ToResponse(),
	})
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "User ID is required",
		})
		return
	}

	// Cannot delete yourself
	actorID := c.GetString("userID")
	if actorID == userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Cannot delete your own account",
		})
		return
	}

	// Check role hierarchy: actor must outrank target
	actorRole := c.GetString("role")
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}
	userService, err := h.getTenantService(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to access tenant database: " + err.Error(),
		})
		return
	}
	existingUser, err := userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if existingUser == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}
	if !middleware.CanManageRole(actorRole, existingUser.Role) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Cannot delete user with role '" + existingUser.Role + "': insufficient permissions",
		})
		return
	}

	if err := userService.DeleteUser(c.Request.Context(), userID, actorID, tenantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User deleted successfully",
	})
}

// GetUser gets a user by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "User ID is required",
		})
		return
	}

	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}
	userService, err := h.getTenantService(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to access tenant database: " + err.Error(),
		})
		return
	}

	user, err := userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user.ToResponse(),
	})
}

// ListUsers lists all users with pagination
func (h *UserHandler) ListUsers(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}

	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}
	userService, err := h.getTenantService(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to access tenant database: " + err.Error(),
		})
		return
	}

	result, err := userService.ListUsers(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Convert to response format
	response := make([]models.UserResponse, len(result.Users))
	for i, user := range result.Users {
		response[i] = user.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"users": response,
			"total": result.Total,
			"page":  result.Page,
			"limit": result.Limit,
		},
	})
}

// UnlockUser unlocks a locked user account
func (h *UserHandler) UnlockUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "User ID is required",
		})
		return
	}

	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}
	unlockedBy := c.GetString("userID")
	userService, err := h.getTenantService(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to access tenant database: " + err.Error(),
		})
		return
	}

	if err := userService.UnlockUser(c.Request.Context(), userID, unlockedBy, tenantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User unlocked successfully",
	})
}
