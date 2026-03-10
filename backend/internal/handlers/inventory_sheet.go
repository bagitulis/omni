package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/google"
)

// GetSyncStatus handles POST /api/inventory/sync-status
// Returns sync status for a registered sheet
func (h *InventoryHandler) GetSyncStatus(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	var req SyncStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "sheetId is required",
		})
		return
	}

	if req.SheetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "sheetId is required",
		})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get spreadsheet from registry
	ctx := c.Request.Context()
	var spreadsheet models.Spreadsheet
	if err := db.WithContext(ctx).Where("id = ? AND tenant_id = ?", req.SheetID, tenantID).First(&spreadsheet).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Sheet not found",
		})
		return
	}

	// Check if locked
	isLocked := spreadsheet.LockedBy != nil && *spreadsheet.LockedBy != ""

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"spreadsheet_name": spreadsheet.Name,
			"is_locked":        isLocked,
			"last_used_at":     spreadsheet.LastUsedAt,
			"sync_enabled":     spreadsheet.AutoSync,
			"sync_interval":    spreadsheet.SyncInterval,
		},
	})
}

// ExportToSheet handles POST /api/inventory/export-to-sheet
func (h *InventoryHandler) ExportToSheet(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	var req ExportToSheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	if req.SheetID == "" || len(req.Data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "sheetId and data array are required",
		})
		return
	}

	if h.googleAuth == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Google Sheets integration not configured",
		})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get spreadsheet from registry
	ctx := c.Request.Context()
	var spreadsheet models.Spreadsheet
	if err := db.WithContext(ctx).Where("id = ? AND tenant_id = ?", req.SheetID, tenantID).First(&spreadsheet).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Sheet not found",
		})
		return
	}

	// Check if locked by another user
	if spreadsheet.LockedBy != nil && *spreadsheet.LockedBy != "" {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Sheet is currently locked for editing",
		})
		return
	}

	// Create sheets service
	sheetsService := google.NewSheetsService(h.googleAuth, tenantID)

	// Get headers from first row
	headers := extractHeaders(req.Data[0])

	// Build values array using helper
	values := buildSheetValues(req.Data, headers)

	// Get target sheet name
	targetSheet := "Sheet1"
	if spreadsheet.SheetName != "" {
		targetSheet = spreadsheet.SheetName
	}

	// Clear and write
	if err := sheetsService.ClearRange(c.Request.Context(), spreadsheet.SpreadsheetID, targetSheet+"!A:Z"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to clear sheet before writing: " + err.Error(),
		})
		return
	}

	if err := sheetsService.WriteRange(c.Request.Context(), spreadsheet.SpreadsheetID, targetSheet+"!A1", values); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to write data: " + err.Error(),
		})
		return
	}

	// Update last used
	now := time.Now()
	spreadsheet.LastUsedAt = &now
	db.WithContext(ctx).Save(&spreadsheet)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Exported " + strconv.Itoa(len(req.Data)) + " items successfully",
		"stats": gin.H{
			"rows":    len(req.Data),
			"columns": len(extractHeaders(req.Data[0])),
		},
	})
}

// ImportFromSheet handles POST /api/inventory/import-from-sheet
func (h *InventoryHandler) ImportFromSheet(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	var req ImportFromSheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	if req.SheetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "sheetId is required",
		})
		return
	}

	if h.googleAuth == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Google Sheets integration not configured",
		})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get spreadsheet from registry
	ctx := c.Request.Context()
	var spreadsheet models.Spreadsheet
	if err := db.WithContext(ctx).Where("id = ? AND tenant_id = ?", req.SheetID, tenantID).First(&spreadsheet).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Sheet not found",
		})
		return
	}

	// Create sheets service
	sheetsService := google.NewSheetsService(h.googleAuth, tenantID)

	// Get target sheet name
	targetSheet := req.SheetName
	if targetSheet == "" {
		targetSheet = spreadsheet.SheetName
	}
	if targetSheet == "" {
		targetSheet = "Sheet1"
	}

	// Read data from sheet
	data, err := sheetsService.ReadRange(c.Request.Context(), spreadsheet.SpreadsheetID, targetSheet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to read data: " + err.Error(),
		})
		return
	}

	if len(data) < 2 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
		})
		return
	}

	// Parse headers from first row using helper
	headers := parseSheetHeaders(data[0])

	// Convert rows to maps using helper
	result := convertRowsToMaps(data[1:], headers)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
