package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/inventory"
	"github.com/rs/xid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type updateInventoryConfigRequest struct {
	SpreadsheetID       *string         `json:"spreadsheet_id"`
	SheetName           *string         `json:"sheet_name"`
	SelectedColumns     json.RawMessage `json:"selected_columns"`
	AllColumns          json.RawMessage `json:"all_columns"`
	HeaderRow           *int            `json:"header_row"`
	DataStartRow        *int            `json:"data_start_row"`
	KeyColumn           *string         `json:"key_column"`
	TotalColumn         *string         `json:"total_column"`
	RawTotalColumn      *string         `json:"raw_total_column"`
	AutoColumn          *string         `json:"auto_column"`
	AutoSync            *bool           `json:"auto_sync"`
	SyncIntervalSeconds *int            `json:"sync_interval_seconds"`
	PriceColumn         *string         `json:"price_column"`
	PriceColumnShopee   *string         `json:"price_column_shopee"`
	PriceColumnTiktok   *string         `json:"price_column_tiktok"`
	PriceColumnLazada   *string         `json:"price_column_lazada"`
	ShopeeRatio         *float64        `json:"shopee_ratio"`
	TiktokRatio         *float64        `json:"tiktok_ratio"`
}

// UpdateConfig handles PUT /api/inventory/config
func (h *InventoryHandler) UpdateConfig(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	var req updateInventoryConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	selectedColumns, selectedColumnsProvided, err := parseOptionalColumnsField(req.SelectedColumns)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	allColumns, allColumnsProvided, err := parseOptionalColumnsField(req.AllColumns)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	service := inventory.NewInventoryService(db, tenantID)
	settings, err := service.GetSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if settings == nil {
		settings = &models.InventorySettings{
			TenantID:        tenantID,
			HeaderRow:       1,
			DataStartRow:    2,
			SyncIntervalSec: 300,
		}
	}
	settings.TenantID = tenantID

	applyInventoryConfigUpdates(settings, req, selectedColumns, selectedColumnsProvided, allColumns, allColumnsProvided)

	if err := upsertInventorySettings(c.Request.Context(), db, settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var persistedSettings models.InventorySettings
	if err := db.WithContext(c.Request.Context()).Where("tenant_id = ?", tenantID).First(&persistedSettings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	settings = &persistedSettings

	var responseSelectedColumns []string
	if settings.SelectedColumns != "" {
		if err := json.Unmarshal([]byte(settings.SelectedColumns), &responseSelectedColumns); err != nil {
			responseSelectedColumns = parseCommaSeparatedColumns(settings.SelectedColumns)
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
			SelectedColumns:     responseSelectedColumns,
			HeaderRow:           settings.HeaderRow,
			DataStartRow:        settings.DataStartRow,
			KeyColumn:           settings.KeyColumn,
			TotalColumn:         settings.TotalColumn,
			RawTotalColumn:      settings.RawTotalColumn,
			AutoColumn:          settings.AutoColumn,
			AutoSync:            settings.AutoSync,
			SyncIntervalSeconds: settings.SyncIntervalSec,
			LastSyncTimestamp:   lastSync,
			PriceColumn:         settings.PriceColumn,
			PriceColumnShopee:   settings.PriceColumnShopee,
			PriceColumnTiktok:   settings.PriceColumnTiktok,
			PriceColumnLazada:   settings.PriceColumnLazada,
			ShopeeRatio:         settings.ShopeeRatio,
			TiktokRatio:         settings.TiktokRatio,
		},
	})
}

func applyInventoryConfigUpdates(
	settings *models.InventorySettings,
	req updateInventoryConfigRequest,
	selectedColumns string,
	selectedColumnsProvided bool,
	allColumns string,
	allColumnsProvided bool,
) {
	if req.SpreadsheetID != nil {
		settings.SpreadsheetID = *req.SpreadsheetID
	}
	if req.SheetName != nil {
		settings.SheetName = *req.SheetName
	}
	if req.HeaderRow != nil {
		settings.HeaderRow = *req.HeaderRow
	}
	if req.DataStartRow != nil {
		settings.DataStartRow = *req.DataStartRow
	}
	if req.KeyColumn != nil {
		settings.KeyColumn = *req.KeyColumn
	}
	if req.TotalColumn != nil {
		settings.TotalColumn = *req.TotalColumn
	}
	if req.RawTotalColumn != nil {
		settings.RawTotalColumn = *req.RawTotalColumn
	}
	if req.AutoColumn != nil {
		settings.AutoColumn = *req.AutoColumn
	}
	if req.AutoSync != nil {
		settings.AutoSync = *req.AutoSync
	}
	if req.SyncIntervalSeconds != nil {
		settings.SyncIntervalSec = *req.SyncIntervalSeconds
	}
	if req.PriceColumn != nil {
		settings.PriceColumn = *req.PriceColumn
	}
	if req.PriceColumnShopee != nil {
		settings.PriceColumnShopee = *req.PriceColumnShopee
	}
	if req.PriceColumnTiktok != nil {
		settings.PriceColumnTiktok = *req.PriceColumnTiktok
	}
	if req.PriceColumnLazada != nil {
		settings.PriceColumnLazada = *req.PriceColumnLazada
	}
	if req.ShopeeRatio != nil {
		settings.ShopeeRatio = *req.ShopeeRatio
	}
	if req.TiktokRatio != nil {
		settings.TiktokRatio = *req.TiktokRatio
	}
	if selectedColumnsProvided {
		settings.SelectedColumns = selectedColumns
	}
	if allColumnsProvided {
		settings.AllColumns = allColumns
	}
}

func parseOptionalColumnsField(raw json.RawMessage) (string, bool, error) {
	if len(raw) == 0 {
		return "", false, nil
	}

	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "null" {
		return "", true, nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err == nil {
		return value, true, nil
	}

	var columns []string
	if err := json.Unmarshal(raw, &columns); err == nil {
		encoded, marshalErr := json.Marshal(columns)
		if marshalErr != nil {
			return "", true, fmt.Errorf("failed to encode column values: %w", marshalErr)
		}
		return string(encoded), true, nil
	}

	return "", true, fmt.Errorf("selected_columns and all_columns must be string or string array")
}

func upsertInventorySettings(ctx context.Context, db *gorm.DB, settings *models.InventorySettings) error {
	if strings.TrimSpace(settings.ID) == "" {
		settings.ID = xid.New().String()
	}
	settings.UpdatedAt = time.Now()

	return db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "tenant_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"spreadsheet_id":        settings.SpreadsheetID,
				"sheet_name":            settings.SheetName,
				"selected_columns":      settings.SelectedColumns,
				"all_columns":           settings.AllColumns,
				"header_row":            settings.HeaderRow,
				"data_start_row":        settings.DataStartRow,
				"key_column":            settings.KeyColumn,
				"total_column":          settings.TotalColumn,
				"raw_total_column":      settings.RawTotalColumn,
				"auto_column":           settings.AutoColumn,
				"auto_sync":             settings.AutoSync,
				"sync_interval_seconds": settings.SyncIntervalSec,
				"price_column":          settings.PriceColumn,
				"price_column_shopee":   settings.PriceColumnShopee,
				"price_column_tiktok":   settings.PriceColumnTiktok,
				"price_column_lazada":   settings.PriceColumnLazada,
				"shopee_ratio":          settings.ShopeeRatio,
				"tiktok_ratio":          settings.TiktokRatio,
				"updated_at":            settings.UpdatedAt,
			}),
		}).
		Create(settings).Error
}
