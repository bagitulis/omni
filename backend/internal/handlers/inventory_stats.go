package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
)

// InventoryStats represents inventory statistics response
type InventoryStats struct {
	TotalRecords int               `json:"total_records"`
	TotalColumns int               `json:"total_columns"`
	Columns      []InventoryColumn `json:"columns"`
	LastSync     *string           `json:"last_sync"`
	DBSizeKB     int               `json:"db_size_kb"`
}

// InventoryColumn represents column metadata
type InventoryColumn struct {
	ColumnName     string `json:"column_name"`
	ColumnType     string `json:"column_type"`
	IsKey          bool   `json:"is_key"`
	SpreadsheetCol string `json:"spreadsheet_column"`
	ColumnPosition int    `json:"column_position"`
}

// GetStats handles GET /api/inventory/stats
func (h *InventoryHandler) GetStats(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Always count records directly from inventory_records table
	var count int64
	db.WithContext(ctx).Model(&models.InventoryRecord{}).Where("tenant_id = ?", tenantID).Count(&count)

	var selectedColumns []string
	var keyColumn string
	var lastSync *string

	// Try to get settings from inventory_settings first
	var settings models.InventorySettings
	if err := db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&settings).Error; err == nil {
		// Found inventory_settings
		if settings.SelectedColumns != "" {
			if err := json.Unmarshal([]byte(settings.SelectedColumns), &selectedColumns); err != nil {
				selectedColumns = parseCommaSeparatedColumns(settings.SelectedColumns)
			}
		}
		keyColumn = settings.KeyColumn
		if settings.LastSyncTimestamp != nil {
			ts := settings.LastSyncTimestamp.Format("2006-01-02T15:04:05Z07:00")
			lastSync = &ts
		}
	} else {
		// Fallback: get columns from first inventory record if exists
		if count > 0 {
			var firstRecord models.InventoryRecord
			if err := db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&firstRecord).Error; err == nil {
				// Parse Data JSONB to extract column names
				if firstRecord.Data != "" {
					var columnData map[string]interface{}
					if err := json.Unmarshal([]byte(firstRecord.Data), &columnData); err == nil {
						for colName := range columnData {
							selectedColumns = append(selectedColumns, colName)
						}
					}
				}
				keyColumn = firstRecord.KeyColumnName
			}
		}
	}

	columns := make([]InventoryColumn, 0)
	for i, col := range selectedColumns {
		columns = append(columns, InventoryColumn{
			ColumnName:     col,
			ColumnType:     "text",
			IsKey:          col == keyColumn,
			SpreadsheetCol: string(rune('A' + i)),
			ColumnPosition: i + 1,
		})
	}

	// Return both status: SUCCESS (legacy) and success: true (new) for frontend compatibility
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": InventoryStats{
			TotalRecords: int(count),
			TotalColumns: len(selectedColumns),
			Columns:      columns,
			LastSync:     lastSync,
			DBSizeKB:     0,
		},
	})
}

// GetPlatformStatus handles GET /api/inventory/platform-status
// Returns real stock/price data from platform staging tables (shopee_skus, tiktok_skus, lazada_skus).
func (h *InventoryHandler) GetPlatformStatus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	type PlatformStockSummary struct {
		Platform   string `json:"platform"`
		TotalSkus  int64  `json:"total_skus"`
		TotalStock int64  `json:"total_stock"`
		AvgPrice   float64 `json:"avg_price"`
	}

	results := make([]PlatformStockSummary, 0, 3)

	// Query Shopee staging
	var shopeeSummary struct {
		Count    int64
		SumStock int64
		AvgPrice float64
	}
	db.WithContext(ctx).Table("shopee_skus").
		Where("tenant_id = ?", tenantID).
		Select("COUNT(*) as count, COALESCE(SUM(quantity), 0) as sum_stock, COALESCE(AVG(price), 0) as avg_price").
		Scan(&shopeeSummary)
	if shopeeSummary.Count > 0 {
		results = append(results, PlatformStockSummary{
			Platform:   "shopee",
			TotalSkus:  shopeeSummary.Count,
			TotalStock: shopeeSummary.SumStock,
			AvgPrice:   shopeeSummary.AvgPrice,
		})
	}

	// Query TikTok staging
	var tiktokSummary struct {
		Count    int64
		SumStock int64
		AvgPrice float64
	}
	db.WithContext(ctx).Table("tiktok_skus").
		Where("tenant_id = ?", tenantID).
		Select("COUNT(*) as count, COALESCE(SUM(quantity), 0) as sum_stock, COALESCE(AVG(price), 0) as avg_price").
		Scan(&tiktokSummary)
	if tiktokSummary.Count > 0 {
		results = append(results, PlatformStockSummary{
			Platform:   "tiktok",
			TotalSkus:  tiktokSummary.Count,
			TotalStock: tiktokSummary.SumStock,
			AvgPrice:   tiktokSummary.AvgPrice,
		})
	}

	// Query Lazada staging
	var lazadaSummary struct {
		Count    int64
		SumStock int64
		AvgPrice float64
	}
	db.WithContext(ctx).Table("lazada_skus").
		Where("tenant_id = ?", tenantID).
		Select("COUNT(*) as count, COALESCE(SUM(quantity), 0) as sum_stock, COALESCE(AVG(price), 0) as avg_price").
		Scan(&lazadaSummary)
	if lazadaSummary.Count > 0 {
		results = append(results, PlatformStockSummary{
			Platform:   "lazada",
			TotalSkus:  lazadaSummary.Count,
			TotalStock: lazadaSummary.SumStock,
			AvgPrice:   lazadaSummary.AvgPrice,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   len(results),
		"results": results,
	})
}
