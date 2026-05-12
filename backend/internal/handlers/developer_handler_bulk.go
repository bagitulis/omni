package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/repositories"
)

const maxBulkUsers = 50

// bulkResetPasswordRequest represents the request body for bulk password reset.
type bulkResetPasswordRequest struct {
	UserIDs     []string `json:"user_ids" binding:"required"`
	TenantID    string   `json:"tenant_id" binding:"required"`
	NewPassword string   `json:"new_password" binding:"required"`
}

// bulkDisableUsersRequest represents the request body for bulk user disable.
type bulkDisableUsersRequest struct {
	UserIDs  []string `json:"user_ids" binding:"required"`
	TenantID string   `json:"tenant_id" binding:"required"`
}

// bulkErrorDetail represents a single failure in a bulk operation.
type bulkErrorDetail struct {
	UserID string `json:"user_id"`
	Error  string `json:"error"`
}

// bulkOperationResult represents the result of a bulk operation.
type bulkOperationResult struct {
	SuccessCount int               `json:"success_count"`
	FailureCount int               `json:"failure_count"`
	Errors       []bulkErrorDetail `json:"errors"`
}

// BulkResetPasswords resets passwords for multiple users in a tenant.
// POST /api/dev/users/bulk-reset-password
func (h *DeveloperHandler) BulkResetPasswords(c *gin.Context) {
	var req bulkResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation failed: " + err.Error(),
		})
		return
	}

	if len(req.UserIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "user_ids must not be empty",
		})
		return
	}

	if len(req.UserIDs) > maxBulkUsers {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("user_ids exceeds maximum of %d users per request", maxBulkUsers),
		})
		return
	}

	// Verify tenant exists
	_, err := h.tenantService.GetTenantDB(req.TenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid tenant_id: " + err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	resetBy := c.GetString("userID")
	if resetBy == "" {
		resetBy = "developer_bulk"
	}

	result := bulkOperationResult{
		Errors: make([]bulkErrorDetail, 0),
	}

	for _, userID := range req.UserIDs {
		err := h.userMgmtService.ResetPassword(ctx, userID, req.NewPassword, resetBy, req.TenantID)
		if err != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, bulkErrorDetail{
				UserID: userID,
				Error:  err.Error(),
			})
		} else {
			result.SuccessCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// BulkDisableUsers disables multiple users in a tenant by locking their accounts.
// POST /api/dev/users/bulk-disable
func (h *DeveloperHandler) BulkDisableUsers(c *gin.Context) {
	var req bulkDisableUsersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "validation failed: " + err.Error(),
		})
		return
	}

	if len(req.UserIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "user_ids must not be empty",
		})
		return
	}

	if len(req.UserIDs) > maxBulkUsers {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("user_ids exceeds maximum of %d users per request", maxBulkUsers),
		})
		return
	}

	// Verify tenant exists and get DB
	db, err := h.tenantService.GetTenantDB(req.TenantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid tenant_id: " + err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	repo := repositories.NewUserRepository(db)
	// Lock accounts permanently (100 years) to effectively disable them
	permanentLock := 100 * 365 * 24 * time.Hour

	result := bulkOperationResult{
		Errors: make([]bulkErrorDetail, 0),
	}

	for _, userID := range req.UserIDs {
		// Verify user exists
		user, findErr := repo.FindByID(ctx, userID)
		if findErr != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, bulkErrorDetail{
				UserID: userID,
				Error:  "user not found",
			})
			continue
		}

		// Skip already locked users
		if user.IsLocked() {
			result.SuccessCount++
			continue
		}

		lockErr := repo.LockAccount(ctx, userID, permanentLock)
		if lockErr != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, bulkErrorDetail{
				UserID: userID,
				Error:  "failed to disable user: " + lockErr.Error(),
			})
		} else {
			result.SuccessCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
