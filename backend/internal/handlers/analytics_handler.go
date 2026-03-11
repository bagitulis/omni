package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services"
	"gorm.io/gorm"
)

// AnalyticsHandler handles analytics endpoints
type AnalyticsHandler struct {
	fallbackDB *gorm.DB
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(db *gorm.DB) *AnalyticsHandler {
	return &AnalyticsHandler{fallbackDB: db}
}

// getDB returns the appropriate database for the current request
// Uses GetTenantDB for proper SQLite and PostgreSQL support
func (h *AnalyticsHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDB(c)
}

// getService creates analytics service with tenant-specific DB
func (h *AnalyticsHandler) getService(c *gin.Context) (*services.AnalyticsService, error) {
	db, err := h.getDB(c)
	if err != nil {
		return nil, err
	}
	repo := repositories.NewAnalyticsRepository(db)
	return services.NewAnalyticsService(repo), nil
}

// GetDashboardSummary gets dashboard summary
func (h *AnalyticsHandler) GetDashboardSummary(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		respondBadRequest(c, "Missing tenant_id")
		return
	}

	service, err := h.getService(c)
	if err != nil {
		respondServiceUnavailable(c, "Database unavailable: "+err.Error())
		return
	}

	// Parse date range (default: last 30 days)
	startDate, endDate := h.parseDateRange(c)

	result, err := service.GetDashboardSummary(c.Request.Context(), tenantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetOrderAnalytics gets order analytics
func (h *AnalyticsHandler) GetOrderAnalytics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		respondBadRequest(c, "Missing tenant_id")
		return
	}

	service, err := h.getService(c)
	if err != nil {
		respondServiceUnavailable(c, "Database unavailable: "+err.Error())
		return
	}

	startDate, endDate := h.parseDateRange(c)

	result, err := service.GetOrderAnalytics(c.Request.Context(), tenantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetRevenueAnalytics gets revenue analytics
func (h *AnalyticsHandler) GetRevenueAnalytics(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		respondBadRequest(c, "Missing tenant_id")
		return
	}

	service, err := h.getService(c)
	if err != nil {
		respondServiceUnavailable(c, "Database unavailable: "+err.Error())
		return
	}

	startDate, endDate := h.parseDateRange(c)

	result, err := service.GetRevenueAnalytics(c.Request.Context(), tenantID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetAnalyticsSettings gets analytics settings
func (h *AnalyticsHandler) GetAnalyticsSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		respondBadRequest(c, "Missing tenant_id")
		return
	}

	service, err := h.getService(c)
	if err != nil {
		respondServiceUnavailable(c, "Database unavailable: "+err.Error())
		return
	}

	platform := c.Query("platform")
	if platform == "" {
		platform = models.PlatformShopee
	}

	settings, err := service.GetSettings(c.Request.Context(), tenantID, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

// UpdateAnalyticsSettings updates analytics settings
func (h *AnalyticsHandler) UpdateAnalyticsSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		respondBadRequest(c, "Missing tenant_id")
		return
	}

	service, err := h.getService(c)
	if err != nil {
		respondServiceUnavailable(c, "Database unavailable: "+err.Error())
		return
	}

	var req UpdateAnalyticsSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	settings := &models.AnalyticsSettings{
		TenantID:          tenantID,
		Platform:          req.Platform,
		PriceColumn:       req.PriceColumn,
		FormulaDeduction:  req.FormulaDeduction,
		FormulaMultiplier: req.FormulaMultiplier,
	}

	if err := service.UpdateSettings(c.Request.Context(), settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Settings updated successfully",
	})
}

// GetEscrowSyncStatus gets escrow sync status
func (h *AnalyticsHandler) GetEscrowSyncStatus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		respondBadRequest(c, "Missing tenant_id")
		return
	}

	service, err := h.getService(c)
	if err != nil {
		respondServiceUnavailable(c, "Database unavailable: "+err.Error())
		return
	}

	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

	result, err := service.GetEscrowSyncStatus(c.Request.Context(), tenantID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// parseDateRange parses date range from query params
func (h *AnalyticsHandler) parseDateRange(c *gin.Context) (time.Time, time.Time) {
	now := time.Now()
	endDate := now
	startDate := now.AddDate(0, 0, -30) // Default: last 30 days

	if startDateStr := c.Query("startDate"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := c.Query("endDate"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed.Add(24*time.Hour - time.Second)
		}
	}

	return startDate, endDate
}
