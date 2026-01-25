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
	ID            string                 `json:"id"`
	KeyValue      string                 `json:"key_value"`
	KeyColumnName string                 `json:"key_column_name"`
	Data          map[string]interface{} `json:"data"`
	CreatedAt     string                 `json:"created_at,omitempty"`
	UpdatedAt     string                 `json:"updated_at,omitempty"`
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
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant_id is required"})
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
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "tenant ID required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "100")
	search := c.Query("search")

	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)
	if limit > 500 {
		limit = 500
	}

	var records []models.InventoryRecord
	query := db.WithContext(c.Request.Context()).Where("tenant_id = ?", tenantID)

	if search != "" {
		query = query.Where("key_value LIKE ? OR data::text LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Model(&models.InventoryRecord{}).Count(&total)

	if err := query.Offset(offset).Limit(limit).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	items := make([]InventoryListItem, 0, len(records))
	for _, rec := range records {
		var data map[string]interface{}
		if rec.Data != "" {
			json.Unmarshal([]byte(rec.Data), &data)
		}
		items = append(items, InventoryListItem{
			ID:            rec.ID,
			KeyValue:      rec.KeyValue,
			KeyColumnName: rec.KeyColumnName,
			Data:          data,
			CreatedAt:     rec.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     rec.UpdatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"total":   total,
		"offset":  offset,
		"limit":   limit,
	})
}
