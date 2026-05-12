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

// bulkUserItem represents a single user with their tenant for bulk operations.
type bulkUserItem struct {
	UserID   string `json:"user_id" binding:"required"`
	TenantID string `json:"tenant_id" binding:"required"`
}

// bulkResetPasswordRequest represents the request body for bulk password reset.
type bulkResetPasswordRequest struct {
	Items       []bulkUserItem `json:"items" binding:"required"`
	NewPassword string         `json:"new_password" binding:"required"`
}

// bulkDisableUsersRequest represents the request body for bulk user disable.
type bulkDisableUsersRequest struct {
	Items []bulkUserItem `json:"items" binding:"required"`
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

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "items must not be empty",
		})
		return
	}

	if len(req.Items) > maxBulkUsers {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("items exceeds maximum of %d users per request", maxBulkUsers),
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

	for _, item := range req.Items {
		// Verify tenant exists for each item
		_, err := h.tenantService.GetTenantDB(item.TenantID)
		if err != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, bulkErrorDetail{
				UserID: item.UserID,
				Error:  "invalid tenant_id: " + err.Error(),
			})
			continue
		}

		err = h.userMgmtService.ResetPassword(ctx, item.UserID, req.NewPassword, resetBy, item.TenantID)
		if err != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, bulkErrorDetail{
				UserID: item.UserID,
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

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "items must not be empty",
		})
		return
	}

	if len(req.Items) > maxBulkUsers {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("items exceeds maximum of %d users per request", maxBulkUsers),
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Lock accounts permanently (100 years) to effectively disable them
	permanentLock := 100 * 365 * 24 * time.Hour

	result := bulkOperationResult{
		Errors: make([]bulkErrorDetail, 0),
	}

	for _, item := range req.Items {
		// Get tenant DB for each item
		db, err := h.tenantService.GetTenantDB(item.TenantID)
		if err != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, bulkErrorDetail{
				UserID: item.UserID,
				Error:  "invalid tenant_id: " + err.Error(),
			})
			continue
		}

		repo := repositories.NewUserRepository(db)

		// Verify user exists
		user, findErr := repo.FindByID(ctx, item.UserID)
		if findErr != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, bulkErrorDetail{
				UserID: item.UserID,
				Error:  "user not found",
			})
			continue
		}

		// Skip already locked users
		if user.IsLocked() {
			result.SuccessCount++
			continue
		}

		lockErr := repo.LockAccount(ctx, item.UserID, permanentLock)
		if lockErr != nil {
			result.FailureCount++
			result.Errors = append(result.Errors, bulkErrorDetail{
				UserID: item.UserID,
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
