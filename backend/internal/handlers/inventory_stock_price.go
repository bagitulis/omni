package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	inventoryService "github.com/omni/backend/internal/services/inventory"
)

// UpdateStockRequest represents stock update request
type UpdateStockRequest struct {
	SKU       string   `json:"sku" binding:"required"`
	Platform  string   `json:"platform"`
	Platforms []string `json:"platforms"` // Array of platforms to update
}

// UpdateStock handles POST /api/inventory/update-stock
// Gets stock from inventory_records and syncs to marketplace platforms
func (h *InventoryHandler) UpdateStock(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant ID required"})
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
	var record models.InventoryRecord
	err = db.WithContext(c.Request.Context()).
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

	// Initialize credential service
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data"
	}
	credService := services.NewCredentialService(dbPath)

	// Use orchestrator to update marketplace platforms
	orchestrator := inventoryService.NewStockUpdateOrchestrator(db, tenantID, credService)
	result, err := orchestrator.UpdateStock(c.Request.Context(), req.SKU, stockValue, platforms)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result, "stock_from_inventory": stockValue})
}

// UpdateStockBatchRequest represents batch stock update request
type UpdateStockBatchRequest struct {
	SKUs      []string `json:"skus" binding:"required"`
	Platform  string   `json:"platform"`
	Platforms []string `json:"platforms"`
}

// UpdateStockBatch handles POST /api/inventory/update-stock-batch
// Gets stock from inventory_records for each SKU and syncs to marketplaces
func (h *InventoryHandler) UpdateStockBatch(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant ID required"})
		return
	}

	var req UpdateStockBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
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

	orchestrator := inventoryService.NewStockUpdateOrchestrator(db, tenantID, credService)
	results := make([]interface{}, 0, len(req.SKUs))

	for _, sku := range req.SKUs {
		// Get stock from inventory_records
		var record models.InventoryRecord
		err := db.WithContext(c.Request.Context()).
			Where("tenant_id = ? AND key_value = ?", tenantID, sku).
			First(&record).Error
		if err != nil {
			results = append(results, gin.H{"sku": sku, "success": false, "error": "SKU not found"})
			continue
		}

		stockValue := inventoryService.GetQuantity(record)
		result, err := orchestrator.UpdateStock(c.Request.Context(), sku, stockValue, platforms)
		if err != nil {
			results = append(results, gin.H{"sku": sku, "success": false, "error": err.Error()})
			continue
		}
		results = append(results, result)
	}

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
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant ID required"})
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
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant ID required"})
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
			continue
		}
		results = append(results, result)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": results})
}
