package inventory

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	inventoryService "github.com/omni/backend/internal/services/inventory"
	"gorm.io/gorm"
)

// SyncHandler handles inventory sync endpoints
type SyncHandler struct {
	db           *gorm.DB
	sheetsClient inventoryService.SheetsClient
}

// NewSyncHandler creates a new inventory sync handler
func NewSyncHandler(db *gorm.DB, sheetsClient inventoryService.SheetsClient) *SyncHandler {
	return &SyncHandler{db: db, sheetsClient: sheetsClient}
}

// TriggerSync handles POST /api/inventory/sync
func (h *SyncHandler) TriggerSync(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	var req struct {
		SpreadsheetID string `json:"spreadsheet_id"`
		SheetName     string `json:"sheet_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// Body is optional — log for debugging but proceed with defaults
		_ = err // intentionally ignored: empty body uses saved settings
	}

	svc := inventoryService.NewSyncService(h.db, tenantID, h.sheetsClient)
	result, err := svc.SyncFromSheets(c.Request.Context(), req.SpreadsheetID, req.SheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetSyncHistory handles GET /api/inventory/sync/history
func (h *SyncHandler) GetSyncHistory(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	svc := inventoryService.NewInventoryService(h.db, tenantID)
	history, err := svc.GetSyncHistory(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": history})
}
