package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
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

// SaveSettings creates or updates settings
func (s *InventoryService) SaveSettings(ctx context.Context, input SettingsInput) error {
	var existing models.InventorySettings
	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", s.tenantID).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		existing = models.InventorySettings{TenantID: s.tenantID}
	} else if err != nil {
		return err
	}

	existing.SpreadsheetID = input.SpreadsheetID
	existing.SheetName = input.SheetName
	existing.KeyColumn = input.KeyColumn
	existing.HeaderRow = input.HeaderRow
	existing.DataStartRow = input.DataStartRow
	existing.AutoSync = input.AutoSync
	existing.SyncIntervalSec = input.SyncIntervalSec

	if input.AllColumns != nil {
		data, err := json.Marshal(input.AllColumns)
		if err != nil {
			return fmt.Errorf("failed to marshal all_columns: %w", err)
		}
		existing.AllColumns = string(data)
	}
	if input.SelectedColumns != nil {
		data, err := json.Marshal(input.SelectedColumns)
		if err != nil {
			return fmt.Errorf("failed to marshal selected_columns: %w", err)
		}
		existing.SelectedColumns = string(data)
	}

	return s.db.WithContext(ctx).Save(&existing).Error
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
