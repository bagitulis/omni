package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/security"
)

// SecurityHandler handles security endpoints
type SecurityHandler struct {
	notificationService *security.NotificationService
}

// NewSecurityHandler creates a new security handler
func NewSecurityHandler(ns *security.NotificationService) *SecurityHandler {
	return &SecurityHandler{notificationService: ns}
}

// GetAlerts handles GET /api/security/alerts
func (h *SecurityHandler) GetAlerts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	filter := security.AlertFilter{
		TenantID: tenantID,
		Type:     c.Query("type"),
		Severity: c.Query("severity"),
		Limit:    50,
	}

	if since := c.Query("since"); since != "" {
		if t, err := time.Parse(time.RFC3339, since); err == nil {
			filter.Since = t
		}
	}

	alerts := h.notificationService.GetAlerts(filter)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"alerts":  alerts,
		"count":   len(alerts),
	})
}

// ReportIssue handles POST /api/security/report
func (h *SecurityHandler) ReportIssue(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	var req struct {
		Type        string                 `json:"type" binding:"required"`
		Description string                 `json:"description" binding:"required"`
		Details     map[string]interface{} `json:"details,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alert := security.SecurityAlert{
		Type:      req.Type,
		Severity:  "medium",
		Message:   req.Description,
		TenantID:  tenantID,
		IPAddress: c.ClientIP(),
		Details:   req.Details,
	}

	recorded := h.notificationService.RecordAlert(alert)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"recorded": recorded,
		"message":  "issue reported",
	})
}

// GetStats handles GET /api/security/stats
func (h *SecurityHandler) GetStats(c *gin.Context) {
	stats := h.notificationService.GetStats()
	c.JSON(http.StatusOK, gin.H{"success": true, "stats": stats})
}
