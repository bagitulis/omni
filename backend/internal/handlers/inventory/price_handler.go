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

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result, "price_from_inventory": priceValue})
}

// UpdatePriceBatch handles POST /api/inventory/update-price-batch
func (h *PriceHandler) UpdatePriceBatch(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	var req struct {
		Items []inventoryService.PriceUpdateItem `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	svc := inventoryService.NewPriceService(h.db, tenantID)
	result, err := svc.UpdatePriceBatch(c.Request.Context(), req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
