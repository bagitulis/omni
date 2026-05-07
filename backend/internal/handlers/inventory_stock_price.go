package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
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

	platforms := resolvePlatforms(req.Platforms, req.Platform)
	orchestrator := newStockOrchestrator(db, tenantID)
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

	orchestrator := newStockOrchestrator(db, tenantID)
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
		// use the request price directly.
		if req.Price != nil {
			priceValue = *req.Price
		} else {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "SKU not found in inventory"})
			return
		}
	} else {
		priceValue = resolvePriceFromInventory(record, req.Price, req.Platforms, req.Platform)
	}

	platforms := resolvePlatforms(req.Platforms, req.Platform)
	orchestrator := newPriceOrchestrator(db, tenantID)

	if len(platforms) > 1 && record.ID != "" {
		allResults, lastErr := updatePriceMultiPlatform(c.Request.Context(), orchestrator, record, req.SKU, platforms, priceValue)
		respondPriceMultiPlatform(c, req.SKU, allResults, priceValue, lastErr)
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

	orchestrator := newPriceOrchestrator(db, tenantID)

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

	results, successCount, failedCount := countBatchResults(rawResults)

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
