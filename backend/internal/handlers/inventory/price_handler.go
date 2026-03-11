package inventory

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services"
	inventoryService "github.com/omni/backend/internal/services/inventory"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// PriceHandler handles inventory price update endpoints
type PriceHandler struct {
	db          *gorm.DB
	credService *services.CredentialService
}

// NewPriceHandler creates a new price handler
func NewPriceHandler(db *gorm.DB) *PriceHandler {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data"
	}
	return &PriceHandler{
		db:          db,
		credService: services.NewCredentialService(dbPath),
	}
}

// UpdatePrice handles POST /api/inventory/update-price
// Gets price from inventory_records and syncs to marketplace platforms
func (h *PriceHandler) UpdatePrice(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
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

	// Get price from inventory_records
	var record models.InventoryRecord
	err := h.db.WithContext(c.Request.Context()).
		Where("tenant_id = ? AND key_value = ?", tenantID, req.SKU).
		First(&record).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "SKU not found in inventory"})
		return
	}

	// Get price value from inventory data
	priceValue := inventoryService.GetPrice(record)

	// Determine platforms to update
	platforms := req.Platforms
	if len(platforms) == 0 && req.Platform != "" {
		platforms = []string{req.Platform}
	}

	// Use orchestrator to update marketplace platforms
	orchestrator := inventoryService.NewPriceUpdateOrchestrator(h.db, tenantID, h.credService)
	result, err := orchestrator.UpdatePrice(c.Request.Context(), req.SKU, priceValue, platforms)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Record marketplace sync history (fire-and-forget)
	recordPriceSyncHistory(c.Request.Context(), h.db, tenantID, req.SKU, result)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result, "price_from_inventory": priceValue})
}

// UpdatePriceBatch handles POST /api/inventory/update-price-batch
func (h *PriceHandler) UpdatePriceBatch(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	var req struct {
		Items []inventoryService.PriceUpdateItem `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Use orchestrator to sync to platforms instead of just updating DB
	orchestrator := inventoryService.NewPriceUpdateOrchestrator(h.db, tenantID, h.credService)

	allResults := make([]map[string]interface{}, 0, len(req.Items))
	totalSuccess := 0
	totalFailed := 0

	for _, item := range req.Items {
		result, err := orchestrator.UpdatePrice(c.Request.Context(), item.SKU, item.Price, item.Platforms)

		itemResult := map[string]interface{}{
			"sku":       item.SKU,
			"price":     item.Price,
			"platforms": make(map[string]interface{}),
		}

		if err != nil {
			itemResult["success"] = false
			itemResult["error"] = err.Error()
			totalFailed++
		} else {
			itemResult["success"] = result.Success

			// Add per-platform details
			for platform, platformResult := range result.Platforms {
				itemResult["platforms"].(map[string]interface{})[platform] = map[string]interface{}{
					"success": platformResult.Success,
					"error":   platformResult.Error,
					"item_id": platformResult.ItemID,
				}

				// Log platform-specific errors
				if !platformResult.Success && platformResult.Error != "" {
					c.Writer.Header().Add("X-Platform-Error", platform+": "+platformResult.Error)
				}
			}

			if result.Success {
				totalSuccess++
			} else {
				totalFailed++
				if len(result.Errors) > 0 {
					itemResult["errors"] = result.Errors
				}
			}
		}

		allResults = append(allResults, itemResult)
	}

	// Record marketplace sync history for batch operation (fire-and-forget)
	recordPriceBatchSyncHistory(c.Request.Context(), h.db, tenantID, req.Items)

	response := gin.H{
		"success": totalFailed == 0,
		"data": gin.H{
			"total":   len(req.Items),
			"success": totalSuccess,
			"failed":  totalFailed,
			"results": allResults,
		},
	}

	c.JSON(http.StatusOK, response)
}

// recordPriceSyncHistory records marketplace sync history for a single price update (fire-and-forget)
func recordPriceSyncHistory(ctx context.Context, db *gorm.DB, tenantID, sku string, result *inventoryService.PriceUpdateOrchestratorResult) {
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
			Operation:    "price_update",
			Status:       status,
			RequestData:  ptrString(string(requestData)),
			ErrorMessage: errorMsg,
		}

		if err := repo.Create(ctx, entry); err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID).
				Str("sku", sku).
				Str("platform", platform).
				Msg("Failed to record price sync history")
		}
	}
}

// recordPriceBatchSyncHistory records marketplace sync history for batch price updates (fire-and-forget)
func recordPriceBatchSyncHistory(ctx context.Context, db *gorm.DB, tenantID string, items []inventoryService.PriceUpdateItem) {
	repo := repositories.NewMarketplaceSyncHistoryRepo(db)

	// For batch operations, record a summary entry per platform involved
	platformMap := make(map[string]int) // platform -> count

	for _, item := range items {
		for _, platform := range item.Platforms {
			platformMap[platform]++
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
			Operation:   "price_update",
			Status:      "success", // Batch handler returns success if no global error
			RequestData: ptrString(string(requestData)),
		}

		if err := repo.Create(ctx, entry); err != nil {
			log.Error().Err(err).
				Str("tenant_id", tenantID).
				Str("platform", platform).
				Msg("Failed to record batch price sync history")
		}
	}
}
