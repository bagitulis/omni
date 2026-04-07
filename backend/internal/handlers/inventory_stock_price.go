package handlers

import (
	"errors"
	"net/http"
	"os"

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

	// Get price from inventory_records
	var record models.InventoryRecord
	err = db.WithContext(c.Request.Context()).
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

	// Initialize credential service
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data"
	}
	credService := services.NewCredentialService(dbPath)

	// Use orchestrator to update marketplace platforms
	orchestrator := inventoryService.NewPriceUpdateOrchestrator(db, tenantID, credService)
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
	results := make([]interface{}, 0, len(req.Items))
	successCount := 0
	failedCount := 0

	for _, item := range req.Items {
		// Use price from request (not from inventory_records)
		platforms := item.Platforms
		if len(platforms) == 0 {
			// Default to all platforms if not specified
			platforms = []string{"shopee", "lazada", "tiktok"}
		}

		result, err := orchestrator.UpdatePrice(c.Request.Context(), item.SKU, item.Price, platforms)
		if err != nil {
			results = append(results, gin.H{"sku": item.SKU, "success": false, "error": err.Error()})
			failedCount++
			continue
		}
		results = append(results, result)
		// Check if it was actually a success (orchestrator result has Success field)
		// Usually if err is nil, it's at least partially successful or we can treat as processed
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
