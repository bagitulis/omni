package handlers

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	inventoryService "github.com/omni/backend/internal/services/inventory"
)

// UpdateStockRequest represents stock update request
type UpdateStockRequest struct {
	SKU       string   `json:"sku" binding:"required"`
	Stock     *int     `json:"stock,omitempty"`
	Platform  string   `json:"platform"`
	Platforms []string `json:"platforms"` // Array of platforms to update
}

// UpdateStock handles POST /api/inventory/update-stock
// Gets stock from inventory_records and syncs to marketplace platforms
func (h *InventoryHandler) UpdateStock(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	var req UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Get stock from inventory_records
	// Determine platforms to update
	platforms := req.Platforms
	if len(platforms) == 0 && req.Platform != "" {
		platforms = []string{req.Platform}
	}

	// Initialize credential service
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data"
	}
	credService := services.NewCredentialService(dbPath)

	// Use orchestrator to update marketplace platforms
	orchestrator := inventoryService.NewStockUpdateOrchestrator(db, tenantID, credService)
	result, stockValue, err := orchestrator.UpdateStockFromInventory(c.Request.Context(), req.SKU, req.Stock, platforms)
	if err != nil {
		// Fallback: if no inventory record but stock value is provided,
		// sync directly to platform (e.g. fallback-SKU products from staging import).
		if errors.Is(err, inventoryService.ErrInventoryStockSKUNotFound) && req.Stock != nil {
			directResult, directErr := orchestrator.UpdateStock(c.Request.Context(), req.SKU, *req.Stock, platforms)
			if directErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": directErr.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "data": directResult, "stock_from_inventory": *req.Stock})
			return
		}

		if errors.Is(err, inventoryService.ErrInventoryStockSKUNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "SKU not found in inventory"})
			return
		}

		if err.Error() == "stock must be non-negative" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result, "stock_from_inventory": stockValue})
}

// UpdateStockBatchRequest represents batch stock update request
type UpdateStockBatchItem struct {
	SKU       string   `json:"sku" binding:"required"`
	Stock     *int     `json:"stock,omitempty"`
	Platform  string   `json:"platform,omitempty"`
	Platforms []string `json:"platforms,omitempty"`
}

type UpdateStockBatchRequest struct {
	SKUs      []string               `json:"skus"`
	Platform  string                 `json:"platform"`
	Platforms []string               `json:"platforms"`
	Items     []UpdateStockBatchItem `json:"items"`
}

// UpdateStockBatch handles POST /api/inventory/update-stock-batch
// Gets stock from inventory_records for each SKU and syncs to marketplaces
func (h *InventoryHandler) UpdateStockBatch(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	var req UpdateStockBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if len(req.Items) == 0 && len(req.SKUs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "items or skus is required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Initialize credential service
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data"
	}
	credService := services.NewCredentialService(dbPath)

	orchestrator := inventoryService.NewStockUpdateOrchestrator(db, tenantID, credService)
	items := req.Items
	if len(items) == 0 {
		items = make([]UpdateStockBatchItem, 0, len(req.SKUs))

		for _, sku := range req.SKUs {
			items = append(items, UpdateStockBatchItem{
				SKU:       sku,
				Platform:  req.Platform,
				Platforms: req.Platforms,
			})
		}
	}

	syncItems := make([]inventoryService.StockBatchInventoryItem, 0, len(items))
	for _, item := range items {
		syncItems = append(syncItems, inventoryService.StockBatchInventoryItem{
			SKU:       item.SKU,
			Stock:     item.Stock,
			Platform:  item.Platform,
			Platforms: item.Platforms,
		})
	}

	results := orchestrator.UpdateStockBatchFromInventory(c.Request.Context(), syncItems, req.Platforms, req.Platform)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}

// UpdatePriceRequest represents price update request
type UpdatePriceRequest struct {
	SKU       string   `json:"sku" binding:"required"`
	Price     *float64 `json:"price"`
	Platform  string   `json:"platform"`
	Platforms []string `json:"platforms"`
}

// UpdatePrice handles POST /api/inventory/update-price
// Gets price from inventory_records and syncs to marketplace platforms
func (h *InventoryHandler) UpdatePrice(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	var req UpdatePriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Try to get price from inventory_records first.
	var priceValue float64
	var record models.InventoryRecord
	err = db.WithContext(c.Request.Context()).
		Where("tenant_id = ? AND key_value = ?", tenantID, req.SKU).
		First(&record).Error
	if err != nil {
		// Fallback: if no inventory record but price value is provided in request,
		// use the request price directly (e.g. fallback-SKU products from staging import).
		if req.Price != nil {
			priceValue = *req.Price
		} else {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "SKU not found in inventory"})
			return
		}
	} else {
		// Use per-platform prices if available, otherwise base HARGA
		perPlatformPrices := inventoryService.GetPricePerPlatform(record)
		basePrice := perPlatformPrices["base"]
		
		// For single-platform sync: use that platform's specific price
		// For multi-platform or no platform specified: use base price
		// (the orchestrator will be called once per platform with the same price)
		if len(req.Platforms) == 1 {
			platformKey := req.Platforms[0]
			if pp, ok := perPlatformPrices[platformKey]; ok && pp > 0 {
				priceValue = pp
			} else {
				priceValue = basePrice
			}
		} else if req.Platform != "" {
			if pp, ok := perPlatformPrices[req.Platform]; ok && pp > 0 {
				priceValue = pp
			} else {
				priceValue = basePrice
			}
		} else {
			priceValue = basePrice
		}
		
		// If inventory has price=0 but request provides a price, prefer request price.
		if priceValue == 0 && req.Price != nil {
			priceValue = *req.Price
		}
	}

	// Determine platforms to update
	platforms := req.Platforms
	if len(platforms) == 0 && req.Platform != "" {
		platforms = []string{req.Platform}
	}

	// Initialize credential service
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data"
	}
	credService := services.NewCredentialService(dbPath)

	// Use orchestrator to update marketplace platforms
	// For multi-platform sync with per-platform prices, call orchestrator per platform
	orchestrator := inventoryService.NewPriceUpdateOrchestrator(db, tenantID, credService)
	
	if len(platforms) > 1 && record.ID != "" {
		// Multi-platform: use per-platform prices from inventory
		perPlatformPrices := inventoryService.GetPricePerPlatform(record)
		allResults := make(map[string]*inventoryService.PlatformPriceResult)
		var lastErr error
		for _, platform := range platforms {
			pp := perPlatformPrices[platform]
			if pp == 0 {
				pp = priceValue // fallback to resolved price
			}
			result, err := orchestrator.UpdatePrice(c.Request.Context(), req.SKU, pp, []string{platform})
			if err != nil {
				lastErr = err
				continue
			}
			for k, v := range result.Platforms {
				allResults[k] = v
			}
		}
		if len(allResults) == 0 && lastErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": lastErr.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"sku": req.SKU, "platforms": allResults}, "price_from_inventory": priceValue})
		return
	}
	
	// Single platform or no inventory record: use resolved priceValue
	result, err := orchestrator.UpdatePrice(c.Request.Context(), req.SKU, priceValue, platforms)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result, "price_from_inventory": priceValue})
}

// UpdatePriceBatchRequest represents batch price update request
// Updated to match Node.js format: { items: [{ sku, price, platforms? }] }
type UpdatePriceBatchRequest struct {
	Items []struct {
		SKU       string   `json:"sku" binding:"required"`
		Price     float64  `json:"price" binding:"required"`
		Platforms []string `json:"platforms,omitempty"`
	} `json:"items" binding:"required"`
}

// UpdatePriceBatch handles POST /api/inventory/update-price-batch
// Gets price from request (not from inventory_records) and syncs to marketplaces
func (h *InventoryHandler) UpdatePriceBatch(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	var req UpdatePriceBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Initialize credential service
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data"
	}
	credService := services.NewCredentialService(dbPath)

	orchestrator := inventoryService.NewPriceUpdateOrchestrator(db, tenantID, credService)

	// Build batch work items
	items := make([]inventoryService.BatchWorkItem, len(req.Items))
	for i, item := range req.Items {
		platforms := item.Platforms
		if len(platforms) == 0 {
			platforms = []string{"shopee", "lazada", "tiktok"}
		}
		items[i] = inventoryService.BatchWorkItem{
			Index:     i,
			SKU:       item.SKU,
			Platforms: platforms,
		}
	}

	// Execute in parallel with concurrency limit of 3 (external API rate limits)
	config := inventoryService.ParallelBatchConfig{
		MaxConcurrency: 3,
		MaxPerPlatform: 2,
		ItemTimeout:    30 * time.Second,
	}

	rawResults := inventoryService.RunParallelBatch(c.Request.Context(), items, config, func(ctx context.Context, item inventoryService.BatchWorkItem) interface{} {
		result, err := orchestrator.UpdatePrice(ctx, item.SKU, req.Items[item.Index].Price, item.Platforms)
		if err != nil {
			return gin.H{"sku": item.SKU, "success": false, "error": err.Error()}
		}
		return result
	})

	// Count successes/failures from results
	results := make([]interface{}, 0, len(rawResults))
	successCount := 0
	failedCount := 0
	for _, r := range rawResults {
		results = append(results, r)
		if m, ok := r.(map[string]interface{}); ok {
			if success, exists := m["success"]; exists && success == false {
				failedCount++
				continue
			}
		}
		successCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total":   len(req.Items),
			"success": successCount,
			"failed":  failedCount,
			"results": results,
		},
	})
}
