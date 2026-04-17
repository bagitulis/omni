package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SettingsInput represents input for settings update
type SettingsInput struct {
	SpreadsheetID   string   `json:"spreadsheet_id"`
	SheetName       string   `json:"sheet_name"`
	AllColumns      []string `json:"all_columns"`
	SelectedColumns []string `json:"selected_columns"`
	KeyColumn       string   `json:"key_column"`
	HeaderRow       int      `json:"header_row"`
	DataStartRow    int      `json:"data_start_row"`
	AutoSync        bool     `json:"auto_sync"`
	SyncIntervalSec int      `json:"sync_interval_seconds"`
}

// GetSettings retrieves inventory settings
func (s *InventoryService) GetSettings(ctx context.Context) (*models.InventorySettings, error) {
	var settings models.InventorySettings
	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", s.tenantID).
		First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &settings, err
}

// UpdateSettings updates inventory settings
func (s *InventoryService) UpdateSettings(ctx context.Context, settings *models.InventorySettings) error {
	settings.TenantID = s.tenantID
	settings.UpdatedAt = time.Now()
	return s.db.WithContext(ctx).Save(settings).Error
}

// SaveSettings creates or updates settings using atomic upsert.
// Uses ON CONFLICT on the primary key 'id' for a single-SQL atomic operation.
// This eliminates the race condition where concurrent SELECT-then-INSERT
// could cause duplicate key violations on inventory_settings_pkey.
func (s *InventoryService) SaveSettings(ctx context.Context, input SettingsInput) error {
	var allColumnsJSON, selectedColumnsJSON string

	if input.AllColumns != nil {
		data, err := json.Marshal(input.AllColumns)
		if err != nil {
			return fmt.Errorf("failed to marshal all_columns: %w", err)
		}
		allColumnsJSON = string(data)
	}
	if input.SelectedColumns != nil {
		data, err := json.Marshal(input.SelectedColumns)
		if err != nil {
			return fmt.Errorf("failed to marshal selected_columns: %w", err)
		}
		selectedColumnsJSON = string(data)
	}

	settings := models.InventorySettings{
		ID:              "settings",
		TenantID:        s.tenantID,
		SpreadsheetID:   input.SpreadsheetID,
		SheetName:       input.SheetName,
		KeyColumn:       input.KeyColumn,
		HeaderRow:       input.HeaderRow,
		DataStartRow:    input.DataStartRow,
		AutoSync:        input.AutoSync,
		SyncIntervalSec: input.SyncIntervalSec,
		AllColumns:      allColumnsJSON,
		SelectedColumns: selectedColumnsJSON,
		UpdatedAt:       time.Now(),
	}

	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"spreadsheet_id", "sheet_name", "key_column",
				"header_row", "data_start_row", "auto_sync",
				"sync_interval_seconds", "all_columns", "selected_columns",
				"updated_at",
			}),
		}).Create(&settings).Error
}

// GetSyncHistory retrieves sync history
func (s *InventoryService) GetSyncHistory(ctx context.Context, limit int) ([]models.InventorySyncHistory, error) {
	if limit == 0 {
		limit = 20
	}
	var history []models.InventorySyncHistory
	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", s.tenantID).
		Order("synced_at DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

// RecordSyncHistory records a sync operation
func (s *InventoryService) RecordSyncHistory(ctx context.Context, history *models.InventorySyncHistory) error {
	history.TenantID = s.tenantID
	if history.ID == "" {
		history.ID = generateUUID()
	}
	if history.SyncedAt.IsZero() {
		history.SyncedAt = time.Now()
	}
	if history.CreatedAt.IsZero() {
		history.CreatedAt = time.Now()
	}
	return s.db.WithContext(ctx).Create(history).Error
}
