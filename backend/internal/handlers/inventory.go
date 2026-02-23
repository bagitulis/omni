package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/google"
	"github.com/omni/backend/internal/services/inventory"
	"gorm.io/gorm"
)

// InventoryHandler handles inventory endpoints
type InventoryHandler struct {
	fallbackDB *gorm.DB
	googleAuth *google.AuthService
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(db *gorm.DB) *InventoryHandler {
	return &InventoryHandler{fallbackDB: db}
}

// NewInventoryHandlerWithGoogle creates a new inventory handler with Google integration
func NewInventoryHandlerWithGoogle(db *gorm.DB, googleAuth *google.AuthService) *InventoryHandler {
	return &InventoryHandler{fallbackDB: db, googleAuth: googleAuth}
}

// getDB returns the appropriate database for the current request
func (h *InventoryHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// InventoryConfig represents inventory configuration response
type InventoryConfig struct {
	SpreadsheetID       string   `json:"spreadsheet_id"`
	SheetName           string   `json:"sheet_name"`
	SelectedColumns     []string `json:"selected_columns"`
	HeaderRow           int      `json:"header_row"`
	DataStartRow        int      `json:"data_start_row"`
	KeyColumn           string   `json:"key_column"`
	AutoSync            bool     `json:"auto_sync"`
	SyncIntervalSeconds int      `json:"sync_interval_seconds"`
	LastSyncTimestamp   *string  `json:"last_sync_timestamp"`
}

// InventoryListItem represents a single inventory item with JSONB data
type InventoryListItem struct {
	ID             string                         `json:"id"`
	KeyValue       string                         `json:"key_value"`
	KeyColumnName  string                         `json:"key_column_name"`
	Data           map[string]interface{}         `json:"data"`
	SyncStatus     string                         `json:"sync_status"`
	PlatformStatus []inventory.PlatformStatusItem `json:"platform_status"`
	CreatedAt      string                         `json:"created_at,omitempty"`
	UpdatedAt      string                         `json:"updated_at,omitempty"`
}

// parseCommaSeparatedColumns parses a comma-separated string into a slice of column names
func parseCommaSeparatedColumns(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// extractSpreadsheetIDFromURL extracts spreadsheet ID from a full Google Sheets URL
// This is the canonical version used across all inventory handlers
func extractSpreadsheetIDFromURL(urlOrID string) string {
	if urlOrID == "" {
		return ""
	}
	// If it doesn't contain "docs.google.com", assume it's already a plain ID
	if !strings.Contains(urlOrID, "docs.google.com") {
		return urlOrID
	}
	// Extract ID from URL pattern /d/{spreadsheetId}/
	re := regexp.MustCompile(`/d/([a-zA-Z0-9_-]+)`)
	matches := re.FindStringSubmatch(urlOrID)
	if len(matches) >= 2 {
		return matches[1]
	}
	return urlOrID
}

// GetConfig handles GET /api/inventory/config
func (h *InventoryHandler) GetConfig(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var settings models.InventorySettings
	inventorySettingsFound := db.WithContext(c.Request.Context()).Where("tenant_id = ?", tenantID).First(&settings).Error == nil

	// If inventory_settings not found, try to get config from google_sheets_settings
	if !inventorySettingsFound {
		var gsSettings models.GoogleSheetsSettings
		if err := db.WithContext(c.Request.Context()).First(&gsSettings).Error; err == nil {
			// Return config from google_sheets_settings if inventory link is set
			spreadsheetID := extractSpreadsheetIDFromURL(gsSettings.InventorySpreadsheetID)
			sheetName := gsSettings.InventorySheetName

			if spreadsheetID != "" {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data": InventoryConfig{
						SpreadsheetID:       spreadsheetID,
						SheetName:           sheetName,
						SelectedColumns:     []string{},
						HeaderRow:           1,
						DataStartRow:        2,
						KeyColumn:           "",
						AutoSync:            false,
						SyncIntervalSeconds: 300,
						LastSyncTimestamp:   nil,
					},
				})
				return
			}
		}

		// No config found anywhere
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": InventoryConfig{
				SpreadsheetID:       "",
				SheetName:           "",
				SelectedColumns:     []string{},
				HeaderRow:           1,
				DataStartRow:        2,
				KeyColumn:           "",
				AutoSync:            false,
				SyncIntervalSeconds: 300,
				LastSyncTimestamp:   nil,
			},
		})
		return
	}

	var selectedColumns []string
	if settings.SelectedColumns != "" {
		if err := json.Unmarshal([]byte(settings.SelectedColumns), &selectedColumns); err != nil {
			selectedColumns = parseCommaSeparatedColumns(settings.SelectedColumns)
		}
	}

	var lastSync *string
	if settings.LastSyncTimestamp != nil {
		ts := settings.LastSyncTimestamp.Format("2006-01-02T15:04:05Z07:00")
		lastSync = &ts
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": InventoryConfig{
			SpreadsheetID:       settings.SpreadsheetID,
			SheetName:           settings.SheetName,
			SelectedColumns:     selectedColumns,
			HeaderRow:           settings.HeaderRow,
			DataStartRow:        settings.DataStartRow,
			KeyColumn:           settings.KeyColumn,
			AutoSync:            settings.AutoSync,
			SyncIntervalSeconds: settings.SyncIntervalSec,
			LastSyncTimestamp:   lastSync,
		},
	})
}

// GetList handles GET /api/inventory/list
func (h *InventoryHandler) GetList(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Parse query parameters
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "100")
	search := c.Query("search")
	syncStatus := c.QueryArray("sync_status")
	stockStatus := c.Query("stock_status")
	platform := c.QueryArray("platform")

	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)
	if limit > 500 {
		limit = 500
	}

	// Use service layer with platform enrichment
	svc := inventory.NewInventoryService(db, tenantID)
	filter := inventory.ListFilter{
		Search:            search,
		SyncStatus:        syncStatus,
		StockStatus:       stockStatus,
		Platform:          platform,
		LowStockThreshold: 10,
		Limit:             limit,
		Offset:            offset,
	}

	result, platformMap, err := svc.GetRecordsWithPlatformStatus(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Build response items
	items := make([]InventoryListItem, 0, len(result.Data))
	for _, rec := range result.Data {
		var data map[string]interface{}
		if rec.Data != "" {
			json.Unmarshal([]byte(rec.Data), &data)
		}

		platformStatus := platformMap[rec.KeyValue]
		if platformStatus == nil {
			platformStatus = []inventory.PlatformStatusItem{}
		}

		items = append(items, InventoryListItem{
			ID:             rec.ID,
			KeyValue:       rec.KeyValue,
			KeyColumnName:  rec.KeyColumnName,
			Data:           data,
			SyncStatus:     rec.SyncStatus,
			PlatformStatus: platformStatus,
			CreatedAt:      rec.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      rec.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"total":   result.Total,
		"offset":  result.Offset,
		"limit":   limit,
	})
}

// GetSyncHistory handles GET /api/inventory/sync/history
func (h *InventoryHandler) GetSyncHistory(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	svc := inventory.NewInventoryService(db, tenantID)
	history, err := svc.GetSyncHistory(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": history})
}
