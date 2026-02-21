package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/analytics"
	"github.com/omni/backend/internal/services/jobs"
)

// TiktokAnalyticsHandler handles TikTok analytics endpoints
type TiktokAnalyticsHandler struct{}

// NewTiktokAnalyticsHandler creates a new TikTok analytics handler
func NewTiktokAnalyticsHandler() *TiktokAnalyticsHandler {
	return &TiktokAnalyticsHandler{}
}

// getService creates analytics service with tenant context
func (h *TiktokAnalyticsHandler) getService(c *gin.Context) (*analytics.TiktokAnalyticsService, error) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return nil, nil
	}

	// Get tenant-specific database connection
	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get tenant database: " + err.Error()})
		return nil, err
	}

	return analytics.NewTiktokAnalyticsService(tenantDB, tenantID), nil
}

// GetSettings handles GET /api/analytics/tiktok/settings
func (h *TiktokAnalyticsHandler) GetSettings(c *gin.Context) {
	svc, err := h.getService(c)
	if svc == nil {
		return
	}

	settings, err := svc.GetSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": settings})
}

// SaveSettings handles POST /api/analytics/tiktok/settings
func (h *TiktokAnalyticsHandler) SaveSettings(c *gin.Context) {
	svc, _ := h.getService(c)
	if svc == nil {
		return
	}

	var req dto.AnalyticsSettingsDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body"})
		return
	}

	if err := svc.SaveSettings(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Settings saved"})
}

// GetSyncStatus handles GET /api/analytics/tiktok/sync-status
func (h *TiktokAnalyticsHandler) GetSyncStatus(c *gin.Context) {
	svc, err := h.getService(c)
	if svc == nil {
		return
	}

	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

	status, err := svc.GetSyncStatus(c.Request.Context(), month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": status})
}

// SyncEscrow handles POST /api/analytics/tiktok/sync
// Returns immediately with job_id for async background execution
func (h *TiktokAnalyticsHandler) SyncEscrow(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	var req dto.SyncRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		// Fallback to query parameters
		now := time.Now()
		monthStr := c.Query("month")
		yearStr := c.Query("year")

		if monthStr != "" {
			if m, err := strconv.Atoi(monthStr); err == nil {
				req.Month = m
			} else {
				req.Month = int(now.Month())
			}
		} else {
			req.Month = int(now.Month())
		}

		if yearStr != "" {
			if y, err := strconv.Atoi(yearStr); err == nil {
				req.Year = y
			} else {
				req.Year = now.Year()
			}
		} else {
			req.Year = now.Year()
		}
	}

	// Get tenant database
	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get tenant database"})
		return
	}

	// Create job data
	jobData := models.EscrowSyncJobData{
		TenantID:    tenantID,
		Platform:    "tiktok",
		Month:       req.Month,
		Year:        req.Year,
		ForceResync: req.ForceResync,
	}
	jobDataJSON, _ := json.Marshal(jobData)

	// Create queue manager and add job
	qm := jobs.NewQueueManager(tenantDB, tenantID)
	job, err := qm.AddJob(models.CreateJobRequest{
		Type:     models.JobTypeTiktokEscrowSync,
		Data:     string(jobDataJSON),
		Priority: "normal",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create job: " + err.Error()})
		return
	}

	// Return immediately with job_id for async polling
	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"data": gin.H{
			"job_id":  job.ID,
			"month":   req.Month,
			"year":    req.Year,
			"status":  "pending",
			"message": "Escrow sync job created. Poll /api/jobs/{job_id} for progress.",
		},
	})
}

// DeleteSyncData handles DELETE /api/analytics/tiktok/sync
func (h *TiktokAnalyticsHandler) DeleteSyncData(c *gin.Context) {
	svc, _ := h.getService(c)
	if svc == nil {
		return
	}

	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

	if err := svc.DeleteSyncData(c.Request.Context(), month, year); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Sync data deleted"})
}

// GetReconciliation handles GET /api/analytics/tiktok/reconciliation
func (h *TiktokAnalyticsHandler) GetReconciliation(c *gin.Context) {
	svc, err := h.getService(c)
	if svc == nil {
		return
	}

	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

	result, err := svc.GetReconciliation(c.Request.Context(), month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetShippingFeeAnalysis handles GET /api/analytics/tiktok/shipping-fee
func (h *TiktokAnalyticsHandler) GetShippingFeeAnalysis(c *gin.Context) {
	svc, err := h.getService(c)
	if svc == nil {
		return
	}

	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

	result, err := svc.GetShippingFeeAnalysis(c.Request.Context(), month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// RepopulateItems handles POST /api/analytics/tiktok/repopulate-items
// Parses raw_order_data already stored in tiktok_escrow_orders and populates tiktok_escrow_items.
// Use this when escrow_items is empty but orders were already synced.
func (h *TiktokAnalyticsHandler) RepopulateItems(c *gin.Context) {
	svc, err := h.getService(c)
	if svc == nil {
		return
	}

	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

	orders, items, err := svc.RepopulateItems(c.Request.Context(), month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"processed_orders": orders,
			"total_items":      items,
			"month":            month,
			"year":             year,
		},
	})
}

