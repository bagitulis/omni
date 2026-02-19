package inventory

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services"
	inventoryService "github.com/omni/backend/internal/services/inventory"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// StockHandler handles inventory stock update endpoints
type StockHandler struct {
	db          *gorm.DB
	credService *services.CredentialService
}

// NewStockHandler creates a new stock handler
func NewStockHandler(db *gorm.DB) *StockHandler {
	// Initialize credential service for platform API calls
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data" // Default path
	}
	return &StockHandler{
		db:          db,
		credService: services.NewCredentialService(dbPath),
	}
}

// UpdateStock handles POST /api/inventory/update-stock
// Gets stock from inventory_records and syncs to marketplace platforms
func (h *StockHandler) UpdateStock(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	var req struct {
		SKU       string   `json:"sku" binding:"required"`
		Platform  string   `json:"platform"`  // Optional: single platform
		Platforms []string `json:"platforms"` // Optional: array of platforms
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Get stock from inventory_records
	var record models.InventoryRecord
	err := h.db.WithContext(c.Request.Context()).
		Where("tenant_id = ? AND key_value = ?", tenantID, req.SKU).
		First(&record).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "SKU not found in inventory"})
		return
	}

	// Get stock value from inventory data
	stockValue := inventoryService.GetQuantity(record)

	// Determine platforms to update
	platforms := req.Platforms
	if len(platforms) == 0 && req.Platform != "" {
		platforms = []string{req.Platform}
	}

	// Use orchestrator to update marketplace platforms
	orchestrator := inventoryService.NewStockUpdateOrchestrator(h.db, tenantID, h.credService)
	result, err := orchestrator.UpdateStock(c.Request.Context(), req.SKU, stockValue, platforms)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Record marketplace sync history (fire-and-forget)
	recordStockSyncHistory(c.Request.Context(), h.db, tenantID, req.SKU, result)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result, "stock_from_inventory": stockValue})
}

// UpdateStockBatch handles POST /api/inventory/update-stock-batch
func (h *StockHandler) UpdateStockBatch(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	var req struct {
		Items []inventoryService.StockUpdateItem `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	normalizedItems := inventoryService.NormalizeStockUpdateItems(req.Items)
	if len(normalizedItems) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "items is required"})
		return
	}

	svc := inventoryService.NewStockService(h.db, tenantID)
	result, err := svc.UpdateStockBatch(c.Request.Context(), normalizedItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Record marketplace sync history for batch operation (fire-and-forget)
	recordStockBatchSyncHistory(c.Request.Context(), h.db, tenantID, normalizedItems, result)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// ptrString returns a pointer to the given string
func ptrString(s string) *string {
	return &s
}

// recordStockSyncHistory records marketplace sync history for a single stock update (fire-and-forget)
func recordStockSyncHistory(ctx context.Context, db *gorm.DB, tenantID, sku string, result *inventoryService.StockUpdateOrchestratorResult) {
	repo := repositories.NewMarketplaceSyncHistoryRepo(db)

	for platform, platformResult := range result.Platforms {
		status := "failed"
		if platformResult.Success {
			status = "success"
		}

		var errorMsg *string
		if platformResult.Error != "" {
			errorMsg = ptrString(platformResult.Error)
		}

		// Serialize request data
		requestData, _ := json.Marshal(map[string]interface{}{
			"sku":      sku,
			"platform": platform,
		})

		entry := &models.MarketplaceSyncHistory{
			TenantID:     tenantID,
			SKU:          sku,
			Platform:     platform,
			Operation:    "stock_update",
			Status:       status,
			RequestData:  ptrString(string(requestData)),
			ErrorMessage: errorMsg,
		}

		if err := repo.Create(ctx, entry); err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID).
				Str("sku", sku).
				Str("platform", platform).
				Msg("Failed to record stock sync history")
		}
	}
}

// recordStockBatchSyncHistory records marketplace sync history for batch stock updates (fire-and-forget)
func recordStockBatchSyncHistory(ctx context.Context, db *gorm.DB, tenantID string, items []inventoryService.StockUpdateItem, result interface{}) {
	repo := repositories.NewMarketplaceSyncHistoryRepo(db)

	// For batch operations, record a summary entry per platform involved
	platformMap := make(map[string]int) // platform -> count

	for _, item := range items {
		if item.Platform != "" {
			platformMap[item.Platform]++
		}
	}

	// Serialize request data (summary)
	requestData, _ := json.Marshal(map[string]interface{}{
		"batch_size": len(items),
		"platforms":  platformMap,
	})

	for platform := range platformMap {
		entry := &models.MarketplaceSyncHistory{
			TenantID:    tenantID,
			SKU:         "", // Empty for batch summary
			Platform:    platform,
			Operation:   "stock_update",
			Status:      "success", // Batch handler returns success if no global error
			RequestData: ptrString(string(requestData)),
		}

		if err := repo.Create(ctx, entry); err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID).
				Str("platform", platform).
				Msg("Failed to record batch stock sync history")
		}
	}
}

// LookupPlatformIds handles POST /api/inventory/lookup-platform-ids
func (h *StockHandler) LookupPlatformIds(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	var req struct {
		SKUs     []string `json:"skus" binding:"required"`
		Platform string   `json:"platform" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	svc := inventoryService.NewStockService(h.db, tenantID)
	result, err := svc.LookupPlatformIds(c.Request.Context(), req.SKUs, req.Platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
