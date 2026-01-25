package inventory

import (
	"net/http"

	"github.com/gin-gonic/gin"
	inventoryService "github.com/omni/backend/internal/services/inventory"
	"gorm.io/gorm"
)

// SheetHandler handles inventory sheet export/import endpoints
type SheetHandler struct {
	db           *gorm.DB
	sheetsClient inventoryService.SheetWriterClient
}

// NewSheetHandler creates a new sheet handler
func NewSheetHandler(db *gorm.DB, sheetsClient inventoryService.SheetWriterClient) *SheetHandler {
	return &SheetHandler{db: db, sheetsClient: sheetsClient}
}


// ExportToSheet handles POST /api/inventory/export-to-sheet
func (h *SheetHandler) ExportToSheet(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	var req struct {
		SpreadsheetID string   `json:"spreadsheetId" binding:"required"`
		SheetName     string   `json:"sheetName"`
		Columns       []string `json:"columns"`
		Filter        struct {
			Category string `json:"category"`
			LowStock bool   `json:"lowStock"`
		} `json:"filter"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	svc := inventoryService.NewSheetExportService(h.db, tenantID, h.sheetsClient)
	result, err := svc.ExportToSheet(c.Request.Context(), inventoryService.ExportOptions{
		SpreadsheetID: req.SpreadsheetID,
		SheetName:     req.SheetName,
		Fields:        req.Columns,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// ImportFromSheet handles POST /api/inventory/import-from-sheet
func (h *SheetHandler) ImportFromSheet(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	var req struct {
		SpreadsheetID string `json:"spreadsheetId" binding:"required"`
		SheetName     string `json:"sheetName"`
		StartRow      int    `json:"startRow"`
		HasHeader     bool   `json:"hasHeader"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.StartRow == 0 {
		req.StartRow = 1
	}

	svc := inventoryService.NewSheetExportService(h.db, tenantID, h.sheetsClient)
	result, err := svc.ImportFromSheet(c.Request.Context(), inventoryService.ImportOptions{
		SpreadsheetID: req.SpreadsheetID,
		SheetName:     req.SheetName,
		StartRow:      req.StartRow,
		HasHeader:     req.HasHeader,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// SyncFromSheet handles POST /api/inventory/sync-from-sheet
func (h *SheetHandler) SyncFromSheet(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	var req struct {
		SpreadsheetID string `json:"spreadsheetId" binding:"required"`
		SheetName     string `json:"sheetName"`
		Mode          string `json:"mode"` // "merge" or "replace"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if req.Mode == "" {
		req.Mode = "merge"
	}

	svc := inventoryService.NewSheetExportService(h.db, tenantID, h.sheetsClient)
	result, err := svc.SyncFromSheet(c.Request.Context(), inventoryService.SyncOptions{
		SpreadsheetID: req.SpreadsheetID,
		SheetName:     req.SheetName,
		SyncMode:      req.Mode,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// Export handles GET /api/inventory/export (CSV/Excel export)
func (h *SheetHandler) Export(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	format := c.DefaultQuery("format", "csv")

	svc := inventoryService.NewSheetExportService(h.db, tenantID, h.sheetsClient)
	data, contentType, filename, err := svc.Export(c.Request.Context(), format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}

// PartialSync handles POST /api/inventory/sync/partial
func (h *SheetHandler) PartialSync(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	var req struct {
		SKUs          []string `json:"skus" binding:"required"`
		SpreadsheetID string   `json:"spreadsheetId" binding:"required"`
		SheetName     string   `json:"sheetName"`
		KeyColumn     string   `json:"keyColumn"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	svc := inventoryService.NewSheetExportService(h.db, tenantID, h.sheetsClient)
	result, err := svc.PartialSync(c.Request.Context(), req.SKUs, req.SpreadsheetID, req.SheetName, req.KeyColumn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
