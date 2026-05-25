package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/middleware"
	"gorm.io/gorm"
)

// ShopeeAnalyticsService defines the interface for Shopee analytics operations.
// Implementations are provided by the services package (Tasks 5/6).
type ShopeeAnalyticsService interface {
	GetSettings(ctx context.Context, tenantID string) (*dto.AnalyticsSettingsDTO, error)
	SaveSettings(ctx context.Context, tenantID string, settings *dto.AnalyticsSettingsDTO) (*dto.AnalyticsSettingsDTO, error)
	GetSyncStatus(ctx context.Context, tenantID string, month, year int) (*dto.SyncStatusDTO, error)
	SyncEscrow(ctx context.Context, tenantID string, month, year int, forceResync bool) (string, error)
	DeleteSyncData(ctx context.Context, tenantID string, month, year int) error
	GetReconciliation(ctx context.Context, tenantID string, month, year int) (*dto.ReconciliationResultDTO, error)
	GetShippingFeeAnalysis(ctx context.Context, tenantID string, month, year int) (*dto.ShopeeShippingFeeResultDTO, error)
	RepopulateItems(ctx context.Context, tenantID string, period string) error
}

// ShopeeAnalyticsHandler handles Shopee analytics report endpoints.
type ShopeeAnalyticsHandler struct {
	systemDB *gorm.DB
	svc      ShopeeAnalyticsService
}

// NewShopeeAnalyticsHandler creates a new ShopeeAnalyticsHandler.
func NewShopeeAnalyticsHandler(systemDB *gorm.DB) *ShopeeAnalyticsHandler {
	return &ShopeeAnalyticsHandler{systemDB: systemDB}
}

// getService returns the analytics service for the current request context.
// It validates tenant authentication before returning the service.
func (h *ShopeeAnalyticsHandler) getService(c *gin.Context) (ShopeeAnalyticsService, bool) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return nil, false
	}
	if h.svc != nil {
		return h.svc, true
	}
	// Tasks 5/6 will wire concrete service implementation via the svc field.
	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Service not initialized"})
	return nil, false
}

// GetSettings returns analytics settings for the current tenant.
func (h *ShopeeAnalyticsHandler) GetSettings(c *gin.Context) {
	svc, ok := h.getService(c)
	if !ok {
		return
	}
	settings, err := svc.GetSettings(c.Request.Context(), middleware.GetTenantID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": settings})
}

// SaveSettings saves analytics settings for the current tenant.
func (h *ShopeeAnalyticsHandler) SaveSettings(c *gin.Context) {
	svc, ok := h.getService(c)
	if !ok {
		return
	}

	var settings dto.AnalyticsSettingsDTO
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body: " + err.Error()})
		return
	}

	result, err := svc.SaveSettings(c.Request.Context(), middleware.GetTenantID(c), &settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetSyncStatus returns the escrow sync status for the given month/year.
func (h *ShopeeAnalyticsHandler) GetSyncStatus(c *gin.Context) {
	svc, ok := h.getService(c)
	if !ok {
		return
	}

	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

	status, err := svc.GetSyncStatus(c.Request.Context(), middleware.GetTenantID(c), month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": status})
}

// SyncEscrow triggers an escrow sync for the given month/year.
// Accepts parameters from JSON body with GET query fallback.
func (h *ShopeeAnalyticsHandler) SyncEscrow(c *gin.Context) {
	svc, ok := h.getService(c)
	if !ok {
		return
	}

	var req dto.SyncRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		// Fallback to query parameters
		now := time.Now()
		req.Month, _ = strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
		req.Year, _ = strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))
		req.ForceResync = c.Query("force_resync") == "true"
	}

	jobID, err := svc.SyncEscrow(c.Request.Context(), middleware.GetTenantID(c), req.Month, req.Year, req.ForceResync)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"success": true, "data": dto.SyncResultDTO{JobID: jobID}})
}

// DeleteSyncData deletes sync data for the given month/year.
func (h *ShopeeAnalyticsHandler) DeleteSyncData(c *gin.Context) {
	svc, ok := h.getService(c)
	if !ok {
		return
	}

	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

	if err := svc.DeleteSyncData(c.Request.Context(), middleware.GetTenantID(c), month, year); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Sync data deleted successfully"})
}

// GetReconciliation returns reconciliation data for the given month/year.
func (h *ShopeeAnalyticsHandler) GetReconciliation(c *gin.Context) {
	svc, ok := h.getService(c)
	if !ok {
		return
	}

	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

	result, err := svc.GetReconciliation(c.Request.Context(), middleware.GetTenantID(c), month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetShippingFeeAnalysis returns shipping fee analysis for the given month/year.
func (h *ShopeeAnalyticsHandler) GetShippingFeeAnalysis(c *gin.Context) {
	svc, ok := h.getService(c)
	if !ok {
		return
	}

	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))

	result, err := svc.GetShippingFeeAnalysis(c.Request.Context(), middleware.GetTenantID(c), month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// RepopulateItems triggers repopulation of items for the given period.
func (h *ShopeeAnalyticsHandler) RepopulateItems(c *gin.Context) {
	svc, ok := h.getService(c)
	if !ok {
		return
	}

	period := c.DefaultQuery("period", "current")
	if err := svc.RepopulateItems(c.Request.Context(), middleware.GetTenantID(c), period); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Item repopulation triggered successfully"})
}
