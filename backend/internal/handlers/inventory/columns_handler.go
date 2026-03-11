package inventory

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	inventoryService "github.com/omni/backend/internal/services/inventory"
	"gorm.io/gorm"
)

// ColumnsHandler handles inventory column configuration endpoints
type ColumnsHandler struct {
	db *gorm.DB
}

// NewColumnsHandler creates a new columns handler
func NewColumnsHandler(db *gorm.DB) *ColumnsHandler {
	return &ColumnsHandler{db: db}
}

// GetAvailableColumns handles GET /api/inventory/columns/available
func (h *ColumnsHandler) GetAvailableColumns(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	svc := inventoryService.NewColumnsService(h.db, tenantID)
	columns, err := svc.GetAvailableColumns(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": columns})
}

// GetSelectedColumns handles GET /api/inventory/columns/selected
func (h *ColumnsHandler) GetSelectedColumns(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	svc := inventoryService.NewColumnsService(h.db, tenantID)
	columns, err := svc.GetSelectedColumns(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": columns})
}

// SaveSelectedColumns handles POST /api/inventory/columns/selected
func (h *ColumnsHandler) SaveSelectedColumns(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	var req struct {
		Columns []string `json:"columns" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	svc := inventoryService.NewColumnsService(h.db, tenantID)
	if err := svc.SaveSelectedColumns(c.Request.Context(), req.Columns); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Columns saved successfully"})
}
