package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
)

// Handler handles analytics requests
type Handler struct {
	basePath string
}

// NewHandler creates a new analytics handler
func NewHandler(basePath string) *Handler {
	return &Handler{basePath: basePath}
}

// SyncStatusResponse represents the sync status response
type SyncStatusResponse struct {
	LastSync     *string `json:"last_sync"`
	Status       string  `json:"status"`
	TotalRecords int     `json:"total_records"`
	Message      string  `json:"message,omitempty"`
}

// GetShopeeSettings handles GET /api/analytics/shopee/settings
func (h *Handler) GetShopeeSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()
	var settings models.AnalyticsSettings
	result := db.WithContext(ctx).Where("platform = ? AND tenant_id = ?", "shopee", tenantID).First(&settings)
	if result.Error != nil {
		// Return default settings if not found
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"platform":           "shopee",
				"is_enabled":         false,
				"sync_period":        "7d",
				"auto_sync":          false,
				"price_column":       "HARGA",
				"formula_deduction":  1500,
				"formula_multiplier": 0.84,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":                 settings.ID,
			"platform":           settings.Platform,
			"price_column":       settings.PriceColumn,
			"formula_deduction":  settings.FormulaDeduction,
			"formula_multiplier": settings.FormulaMultiplier,
			"is_enabled":         true,
			"sync_period":        "7d",
			"auto_sync":          false,
		},
	})
}

// GetShopeeSyncStatus handles GET /api/analytics/shopee/sync-status
func (h *Handler) GetShopeeSyncStatus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()
	// Check if we have any escrow sync data
	var escrowSync models.ShopeeEscrowSync
	result := db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("synced_at desc").First(&escrowSync)

	resp := SyncStatusResponse{
		Status:       "idle",
		TotalRecords: 0,
	}

	if result.Error == nil {
		lastSync := escrowSync.SyncedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.LastSync = &lastSync
		resp.Status = "synced"
		resp.TotalRecords = escrowSync.TotalOrders
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// GetTiktokSettings handles GET /api/analytics/tiktok/settings
func (h *Handler) GetTiktokSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()
	var settings models.AnalyticsSettings
	result := db.WithContext(ctx).Where("platform = ? AND tenant_id = ?", "tiktok", tenantID).First(&settings)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"platform":           "tiktok",
				"is_enabled":         false,
				"sync_period":        "7d",
				"auto_sync":          false,
				"price_column":       "HARGA",
				"formula_deduction":  1500,
				"formula_multiplier": 0.84,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":                 settings.ID,
			"platform":           settings.Platform,
			"price_column":       settings.PriceColumn,
			"formula_deduction":  settings.FormulaDeduction,
			"formula_multiplier": settings.FormulaMultiplier,
			"is_enabled":         true,
			"sync_period":        "7d",
			"auto_sync":          false,
		},
	})
}

// GetTiktokSyncStatus handles GET /api/analytics/tiktok/sync-status
func (h *Handler) GetTiktokSyncStatus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()
	// Check if we have any TikTok escrow sync data
	var escrowSync models.TiktokEscrowSync
	result := db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").First(&escrowSync)

	resp := SyncStatusResponse{
		Status:       "idle",
		TotalRecords: 0,
	}

	if result.Error == nil {
		lastSync := escrowSync.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.LastSync = &lastSync
		resp.Status = "synced"
		resp.TotalRecords = escrowSync.TotalOrders
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// SyncShopeeRequest represents the sync request body
type SyncShopeeRequest struct {
	Month       int  `json:"month"`
	Year        int  `json:"year"`
	ForceResync bool `json:"force_resync"`
}

// SyncShopeeEscrow handles POST /api/analytics/shopee/sync
// Syncs Shopee escrow data for a specific month
func (h *Handler) SyncShopeeEscrow(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req SyncShopeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request body: month and year required"))
		return
	}

	if req.Month < 1 || req.Month > 12 {
		c.JSON(http.StatusBadRequest, response.Error("Invalid month: must be 1-12"))
		return
	}
	if req.Year < 2020 || req.Year > 2030 {
		c.JSON(http.StatusBadRequest, response.Error("Invalid year"))
		return
	}

	// TODO: Implement actual escrow sync logic
	// For now, return a placeholder response indicating the feature is in development
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Shopee escrow sync started",
		"data": gin.H{
			"month":        req.Month,
			"year":         req.Year,
			"force_resync": req.ForceResync,
			"status":       "pending",
		},
	})
}

// SyncTiktokEscrow handles POST /api/analytics/tiktok/sync
// Syncs TikTok escrow data for a specific month
func (h *Handler) SyncTiktokEscrow(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req SyncShopeeRequest // Same request structure
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request body: month and year required"))
		return
	}

	if req.Month < 1 || req.Month > 12 {
		c.JSON(http.StatusBadRequest, response.Error("Invalid month: must be 1-12"))
		return
	}
	if req.Year < 2020 || req.Year > 2030 {
		c.JSON(http.StatusBadRequest, response.Error("Invalid year"))
		return
	}

	// TODO: Implement actual escrow sync logic
	// For now, return a placeholder response indicating the feature is in development
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "TikTok escrow sync started",
		"data": gin.H{
			"month":        req.Month,
			"year":         req.Year,
			"force_resync": req.ForceResync,
			"status":       "pending",
		},
	})
}
