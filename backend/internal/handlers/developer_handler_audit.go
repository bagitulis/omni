package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// devAuditLogEntry represents an audit log entry with tenant context for cross-tenant viewing.
type devAuditLogEntry struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Action    string `json:"action"`
	Details   string `json:"details"`
	TenantID  string `json:"tenant_id"`
	CreatedAt string `json:"created_at"`
}

// GetAuditLogs returns cross-tenant audit logs for the developer panel.
// GET /api/dev/audit-logs
func (h *DeveloperHandler) GetAuditLogs(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Parse query params
	action := c.Query("action")
	userID := c.Query("user_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Parse date filters
	var startDate, endDate time.Time
	var hasDateFilter bool
	if startDateStr != "" && endDateStr != "" {
		var err error
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid start_date format (use YYYY-MM-DD)",
			})
			return
		}
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid end_date format (use YYYY-MM-DD)",
			})
			return
		}
		endDate = endDate.Add(24*time.Hour - time.Second)
		hasDateFilter = true
	}

	// Get all active tenants
	tenants, err := h.tenantService.GetAvailableTenants(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to list tenants: " + err.Error(),
		})
		return
	}

	// Collect audit logs from all tenants
	var allLogs []models.AuditLog
	for _, t := range tenants {
		db, err := h.tenantService.GetTenantDB(t.ID)
		if err != nil {
			continue
		}

		logs := queryTenantAuditLogs(ctx, db, t.ID, action, userID, startDate, endDate, hasDateFilter)
		allLogs = append(allLogs, logs...)
	}

	// Sort by created_at descending (most recent first)
	sortAuditLogsByDate(allLogs)

	// Apply pagination to aggregated results
	total := len(allLogs)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	pagedLogs := allLogs[start:end]

	// Convert to response format
	entries := make([]devAuditLogEntry, 0, len(pagedLogs))
	for _, log := range pagedLogs {
		entries = append(entries, devAuditLogEntry{
			ID:        log.ID,
			UserID:    log.UserID,
			Action:    log.Action,
			Details:   log.Details,
			TenantID:  log.TenantID,
			CreatedAt: log.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    entries,
		"total":   total,
	})
}

// queryTenantAuditLogs queries audit logs for a single tenant with filters.
func queryTenantAuditLogs(
	ctx context.Context,
	db *gorm.DB,
	tenantID, action, userID string,
	startDate, endDate time.Time,
	hasDateFilter bool,
) []models.AuditLog {
	query := db.WithContext(ctx).Model(&models.AuditLog{}).Where("tenant_id = ?", tenantID)

	if action != "" {
		query = query.Where("action = ?", action)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if hasDateFilter {
		query = query.Where("created_at BETWEEN ? AND ?", startDate, endDate)
	}

	// Limit per tenant to avoid memory issues
	var logs []models.AuditLog
	query.Order("created_at DESC").Limit(200).Find(&logs)
	return logs
}

// sortAuditLogsByDate sorts audit logs by created_at descending.
func sortAuditLogsByDate(logs []models.AuditLog) {
	for i := 1; i < len(logs); i++ {
		for j := i; j > 0 && logs[j].CreatedAt.After(logs[j-1].CreatedAt); j-- {
			logs[j], logs[j-1] = logs[j-1], logs[j]
		}
	}
}
