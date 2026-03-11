package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/google"
	inventoryService "github.com/omni/backend/internal/services/inventory"
)

// SyncFromSheetsRequest represents sync from sheets request
type SyncFromSheetsRequest struct {
	SpreadsheetID string `json:"spreadsheet_id"`
	SheetName     string `json:"sheet_name"`
}

// getInventorySpreadsheetConfig retrieves spreadsheet config from either inventory_settings or google_sheets_settings
func (h *InventoryHandler) getInventorySpreadsheetConfig(c *gin.Context, tenantID string, reqSpreadsheetID, reqSheetName string) (spreadsheetID, sheetName string, err error) {
	db, err := h.getDB(c)
	if err != nil {
		return "", "", err
	}

	// First, try to use request params if provided
	spreadsheetID = reqSpreadsheetID
	sheetName = reqSheetName

	// If not provided in request, try inventory_settings
	if spreadsheetID == "" || sheetName == "" {
		var inventorySettings models.InventorySettings
		if err := db.Where("tenant_id = ?", tenantID).First(&inventorySettings).Error; err == nil {
			if spreadsheetID == "" && inventorySettings.SpreadsheetID != "" {
				spreadsheetID = inventorySettings.SpreadsheetID
			}
			if sheetName == "" && inventorySettings.SheetName != "" {
				sheetName = inventorySettings.SheetName
			}
		}
	}

	// If still not found, try google_sheets_settings (fallback)
	if spreadsheetID == "" || sheetName == "" {
		var gsSettings models.GoogleSheetsSettings
		if err := db.First(&gsSettings).Error; err == nil {
			if spreadsheetID == "" && gsSettings.InventorySpreadsheetID != "" {
				// Extract actual spreadsheet ID from URL if needed
				spreadsheetID = extractSpreadsheetIDFromURL(gsSettings.InventorySpreadsheetID)
			}
			if sheetName == "" && gsSettings.InventorySheetName != "" {
				sheetName = gsSettings.InventorySheetName
			}
		}
	}

	return spreadsheetID, sheetName, nil
}

// SyncFromSheets handles POST /api/inventory/sync/from-sheets
// Syncs inventory data FROM Google Sheets TO database
func (h *InventoryHandler) SyncFromSheets(c *gin.Context) {
	startTime := time.Now()
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "FAILED",
			"message": "Missing tenant_id",
		})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "FAILED",
			"message": err.Error(),
		})
		return
	}

	var req SyncFromSheetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Silently ignore bind errors, will use saved settings
	}

	// Get spreadsheet config from multiple sources (request > inventory_settings > google_sheets_settings)
	spreadsheetID, sheetName, err := h.getInventorySpreadsheetConfig(c, tenantID, req.SpreadsheetID, req.SheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "FAILED",
			"message": err.Error(),
		})
		return
	}

	if spreadsheetID == "" || sheetName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "FAILED",
			"message": "Spreadsheet not configured. Please set spreadsheet link and sheet name in Settings > Google Sheets first.",
		})
		return
	}

	// Check if Google Auth is configured
	if h.googleAuth == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "FAILED",
			"message": "Google Sheets integration not configured. Service account credentials required.",
		})
		return
	}

	// Create sheets service
	sheetsService := google.NewSheetsService(h.googleAuth, tenantID)

	// Create sync service with sheets client
	syncService := inventoryService.NewSyncService(db, tenantID, sheetsService)

	// Execute sync
	result, err := syncService.SyncFromSheets(c.Request.Context(), spreadsheetID, sheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "FAILED",
			"message": err.Error(),
		})
		return
	}

	// Calculate duration
	duration := int(time.Since(startTime).Milliseconds())
	if result.Duration == 0 {
		result.Duration = duration
	}

	// Convert status to lowercase for frontend compatibility
	status := strings.ToLower(result.Status)

	// Return response matching Node.js format (frontend expects lowercase "success")
	c.JSON(http.StatusOK, gin.H{
		"status":            status,
		"message":           result.Message,
		"total_records":     result.TotalRecords,
		"synced_records":    result.NewRecords + result.UpdatedRecords,
		"new_records":       result.NewRecords,
		"updated_records":   result.UpdatedRecords,
		"unchanged_records": result.UnchangedRecords,
		"failed_records":    result.FailedRecords,
		"headers_changed":   result.HeadersChanged,
		"duration":          result.Duration,
		"timestamp":         result.SyncedAt.Format(time.RFC3339),
	})
}

// SyncToSheetsRequest represents sync to sheets request
type SyncToSheetsRequest struct {
	SpreadsheetID string   `json:"spreadsheet_id"`
	SheetName     string   `json:"sheet_name"`
	LockedColumns []string `json:"locked_columns"`
}

// SyncToSheets handles POST /api/inventory/sync/to-sheets
// Syncs inventory data FROM database TO Google Sheets
func (h *InventoryHandler) SyncToSheets(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "FAILED",
			"message": "Missing tenant_id",
		})
		return
	}

	var req SyncToSheetsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Silently ignore bind errors, will use saved settings
	}

	// Get spreadsheet config from multiple sources (request > inventory_settings > google_sheets_settings)
	spreadsheetID, sheetName, err := h.getInventorySpreadsheetConfig(c, tenantID, req.SpreadsheetID, req.SheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "FAILED",
			"message": err.Error(),
		})
		return
	}

	if spreadsheetID == "" || sheetName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "FAILED",
			"message": "Spreadsheet not configured. Please set spreadsheet link and sheet name in Settings > Google Sheets first.",
		})
		return
	}

	// Check if Google Auth is configured
	if h.googleAuth == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "FAILED",
			"message": "Google Sheets integration not configured. Service account credentials required.",
		})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "FAILED",
			"message": err.Error(),
		})
		return
	}

	// Create sheets service
	sheetsService := google.NewSheetsService(h.googleAuth, tenantID)

	// Create sync service with sheets client
	syncService := inventoryService.NewSyncService(db, tenantID, sheetsService)

	// Execute sync to sheets (use extracted config, not raw request)
	result, err := syncService.SyncToSheets(c.Request.Context(), spreadsheetID, sheetName, req.LockedColumns)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "FAILED",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":            strings.ToLower(result.Status),
		"message":           result.Message,
		"total_records":     result.TotalRecords,
		"synced_records":    result.UpdatedRecords,
		"updated_records":   result.UpdatedRecords,
		"unchanged_records": result.UnchangedRecords,
		"new_records":       result.NewRecords,
		"duration":          result.Duration,
		"timestamp":         result.SyncedAt.Format(time.RFC3339),
	})
}
