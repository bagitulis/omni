package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/analytics"
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
	GetSkuOrders(ctx context.Context, tenantID string, sku string, month, year int) (*dto.ShopeeSkuOrdersResultDTO, error)
	GetOrderItems(ctx context.Context, tenantID string, orderSN string, month, year int) (*dto.ShopeeOrderItemsResultDTO, error)
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
	// Create real service from systemDB when available
	if h.systemDB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Service not initialized"})
		return nil, false
	}
	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to get tenant database: " + err.Error()})
		return nil, false
	}
	return analytics.NewShopeeAnalyticsService(h.systemDB, tenantDB, tenantID), true
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

	month, year, ok := parseMonthYear(c)
	if !ok {
		return
	}

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
		month, year, ok := parseMonthYear(c)
		if !ok {
			return
		}
		req.Month = month
		req.Year = year
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

	month, year, ok := parseMonthYear(c)
	if !ok {
		return
	}

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

	month, year, ok := parseMonthYear(c)
	if !ok {
		return
	}

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

	month, year, ok := parseMonthYear(c)
	if !ok {
		return
	}

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

// GetSkuOrders returns orders for a SKU within the given month/year.
func (h *ShopeeAnalyticsHandler) GetSkuOrders(c *gin.Context) {
	svc, ok := h.getService(c)
	if !ok {
		return
	}

	sku := c.Query("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing sku query parameter"})
		return
	}

	month, year, ok := parseMonthYear(c)
	if !ok {
		return
	}

	result, err := svc.GetSkuOrders(c.Request.Context(), middleware.GetTenantID(c), sku, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetOrderItems returns items for an order within the given month/year.
func (h *ShopeeAnalyticsHandler) GetOrderItems(c *gin.Context) {
	svc, ok := h.getService(c)
	if !ok {
		return
	}

	orderSN := c.Query("order_sn")
	if orderSN == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing order_sn query parameter"})
		return
	}

	month, year, ok := parseMonthYear(c)
	if !ok {
		return
	}

	result, err := svc.GetOrderItems(c.Request.Context(), middleware.GetTenantID(c), orderSN, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// parseMonthYear extracts and validates month/year query parameters.
// Returns false with a JSON error response if validation fails.
func parseMonthYear(c *gin.Context) (month, year int, ok bool) {
	now := time.Now()
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(now.Month())))
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid month: must be 1-12"})
		return 0, 0, false
	}
	yearStr := c.DefaultQuery("year", strconv.Itoa(now.Year()))
	year, err = strconv.Atoi(yearStr)
	if err != nil || year < 2000 || year > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid year"})
		return 0, 0, false
	}
	return month, year, true
}
