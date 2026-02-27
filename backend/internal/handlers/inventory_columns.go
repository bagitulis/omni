package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// GetSelectedColumns handles GET /api/inventory/columns/selected
func (h *InventoryHandler) GetSelectedColumns(c *gin.Context) {
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

	ctx := c.Request.Context()
	var selectedColumns []string

	// Try inventory_settings first
	var settings models.InventorySettings
	if err := db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&settings).Error; err == nil {
		if settings.SelectedColumns != "" {
			if err := json.Unmarshal([]byte(settings.SelectedColumns), &selectedColumns); err != nil {
				selectedColumns = parseCommaSeparatedColumns(settings.SelectedColumns)
			}
		}
	}

	// If still empty, try google_sheets_settings
	if len(selectedColumns) == 0 {
		var gsSettings models.GoogleSheetsSettings
		if err := db.WithContext(ctx).First(&gsSettings).Error; err == nil {
			if gsSettings.InventorySelectedColumns != "" {
				if err := json.Unmarshal([]byte(gsSettings.InventorySelectedColumns), &selectedColumns); err != nil {
					selectedColumns = parseCommaSeparatedColumns(gsSettings.InventorySelectedColumns)
				}
			}
		}
	}

	// If still empty, extract from inventory_records
	if len(selectedColumns) == 0 {
		selectedColumns = h.extractColumnsFromRecordsWithContext(ctx, db, tenantID)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"selected_columns": selectedColumns,
	})
}

// UpdateSelectedColumns handles POST /api/inventory/columns/selected
func (h *InventoryHandler) UpdateSelectedColumns(c *gin.Context) {
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

	var req struct {
		SelectedColumns []string `json:"selected_columns" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "selected_columns must be an array"})
		return
	}

	ctx := c.Request.Context()
	columnsJSON, _ := json.Marshal(req.SelectedColumns)

	var settings models.InventorySettings
	findErr := db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&settings).Error

	if findErr == gorm.ErrRecordNotFound {
		settings = models.InventorySettings{
			TenantID:        tenantID,
			SelectedColumns: string(columnsJSON),
		}
		if err := db.WithContext(ctx).Create(&settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
	} else {
		settings.SelectedColumns = string(columnsJSON)
		if err := db.WithContext(ctx).Save(&settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"selected_columns": req.SelectedColumns,
		"message":          "Selected columns updated",
	})
}

// AvailableColumn represents column info for frontend
type AvailableColumn struct {
	Name           string `json:"name"`
	SpreadsheetCol string `json:"spreadsheet_column"`
	Type           string `json:"type"`
	Position       int    `json:"position"`
}

// GetAvailableColumns handles GET /api/inventory/columns/available
func (h *InventoryHandler) GetAvailableColumns(c *gin.Context) {
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

	ctx := c.Request.Context()
	var allColumns []string

	var settings models.InventorySettings
	if err := db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&settings).Error; err != nil {
		// No settings - extract columns from existing inventory_records
		allColumns = h.extractColumnsFromRecordsWithContext(ctx, db, tenantID)
	} else if settings.AllColumns != "" {
		// Parse all_columns from settings (can be JSON array or comma-separated)
		if err := json.Unmarshal([]byte(settings.AllColumns), &allColumns); err != nil {
			// Try comma-separated fallback
			allColumns = parseCommaSeparatedColumns(settings.AllColumns)
		}
	}

	// If still empty, try to extract from records
	if len(allColumns) == 0 {
		allColumns = h.extractColumnsFromRecordsWithContext(ctx, db, tenantID)
	}

	// Build columns with spreadsheet position
	columns := make([]AvailableColumn, 0, len(allColumns))
	for i, colName := range allColumns {
		columns = append(columns, AvailableColumn{
			Name:           colName,
			SpreadsheetCol: columnIndexToLetter(i),
			Type:           "text",
			Position:       i + 1,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"columns": columns,
		"total":   len(columns),
	})
}

// extractColumnsFromRecordsWithContext extracts column names from existing inventory_records JSONB data with context
func (h *InventoryHandler) extractColumnsFromRecordsWithContext(ctx context.Context, db *gorm.DB, tenantID string) []string {
	var record models.InventoryRecord
	if err := db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&record).Error; err != nil {
		return []string{}
	}

	if record.Data == "" {
		return []string{}
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(record.Data), &data); err != nil {
		return []string{}
	}

	// Extract keys and sort them for consistent order
	columns := make([]string, 0, len(data))
	for key := range data {
		columns = append(columns, key)
	}

	// Sort alphabetically for consistent order
	sortColumnsAlphabetically(columns)

	return columns
}

// sortColumnsAlphabetically sorts columns, keeping important columns first
func sortColumnsAlphabetically(columns []string) {
	// Priority columns that should appear first
	priority := map[string]int{
		"Masuk": 1, "TOTAL": 2, "TIKTOK": 3, "Nama Variasi": 4, "EX": 5,
		"HARGA": 6, "LAZADA": 7, "SHOPEE": 8, "UKURAN": 9, "Sisa Stok": 10,
		"MOQ": 11, "PCS": 12, "AUTO": 13, "KARTON": 14, "EXPIRED": 15,
	}

	// Sort with priority first, then alphabetically
	for i := 0; i < len(columns)-1; i++ {
		for j := i + 1; j < len(columns); j++ {
			pi, hasI := priority[columns[i]]
			pj, hasJ := priority[columns[j]]

			swap := false
			if hasI && hasJ {
				swap = pi > pj
			} else if hasJ {
				swap = true
			} else if !hasI && !hasJ {
				swap = columns[i] > columns[j]
			}

			if swap {
				columns[i], columns[j] = columns[j], columns[i]
			}
		}
	}
}

// columnIndexToLetter converts 0-based index to Excel column letter (A, B, ... Z, AA, AB...)
func columnIndexToLetter(index int) string {
	result := ""
	for {
		result = string(rune('A'+index%26)) + result
		index = index/26 - 1
		if index < 0 {
			break
		}
	}
	return result
}
