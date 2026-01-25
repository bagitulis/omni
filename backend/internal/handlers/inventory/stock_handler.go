package inventory

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	inventoryService "github.com/omni/backend/internal/services/inventory"
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

	svc := inventoryService.NewStockService(h.db, tenantID)
	result, err := svc.UpdateStockBatch(c.Request.Context(), req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
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
