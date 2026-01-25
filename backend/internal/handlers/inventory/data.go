package inventory

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	inventoryService "github.com/omni/backend/internal/services/inventory"
	"gorm.io/gorm"
)

// DataHandler handles inventory data endpoints
type DataHandler struct {
	db *gorm.DB
}

// NewDataHandler creates a new inventory data handler
func NewDataHandler(db *gorm.DB) *DataHandler {
	return &DataHandler{db: db}
}

// List handles GET /api/inventory/data
func (h *DataHandler) List(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	filter := inventoryService.ListFilter{
		Search:   c.Query("search"),
		Category: c.Query("category"),
		Platform: c.Query("platform"),
		LowStock: c.Query("low_stock") == "true",
		Limit:    limit,
		Offset:   offset,
	}

	svc := inventoryService.NewInventoryService(h.db, tenantID)
	result, err := svc.GetRecords(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetBySKU handles GET /api/inventory/data/:sku
func (h *DataHandler) GetBySKU(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU required"})
		return
	}

	svc := inventoryService.NewInventoryService(h.db, tenantID)
	record, err := svc.GetBySKU(c.Request.Context(), sku)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if record == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SKU not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": record})
}

// Update handles PUT /api/inventory/data/:sku
func (h *DataHandler) Update(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU required"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prevent updating tenant_id and sku
	delete(updates, "tenantId")
	delete(updates, "sku")
	delete(updates, "id")

	svc := inventoryService.NewInventoryService(h.db, tenantID)
	if err := svc.Update(c.Request.Context(), sku, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Updated"})
}

// Delete handles DELETE /api/inventory/data/:sku
func (h *DataHandler) Delete(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU required"})
		return
	}

	svc := inventoryService.NewInventoryService(h.db, tenantID)
	if err := svc.Delete(c.Request.Context(), sku); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Deleted"})
}

// GetCategories handles GET /api/inventory/categories
func (h *DataHandler) GetCategories(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	svc := inventoryService.NewInventoryService(h.db, tenantID)
	categories, err := svc.GetCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": categories})
}
