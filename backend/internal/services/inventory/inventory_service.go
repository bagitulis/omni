// Package inventory provides inventory management services
package inventory

import (
	"context"
	"encoding/json"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// InventoryService handles inventory operations
type InventoryService struct {
	db       *gorm.DB
	tenantID string
}

// NewInventoryService creates a new inventory service
func NewInventoryService(db *gorm.DB, tenantID string) *InventoryService {
	return &InventoryService{db: db, tenantID: tenantID}
}

// ListFilter represents filter for inventory listing
type ListFilter struct {
	Search   string
	Category string
	Platform string // filter by platform status
	LowStock bool
	Limit    int
	Offset   int
}

// ListResult represents paginated list result
type ListResult struct {
	Total    int64                    `json:"total"`
	Returned int                      `json:"returned"`
	Offset   int                      `json:"offset"`
	Data     []models.InventoryRecord `json:"data"`
}

// GetRecords retrieves inventory records with filtering
func (s *InventoryService) GetRecords(ctx context.Context, filter ListFilter) (*ListResult, error) {
	query := s.db.WithContext(ctx).Where("tenant_id = ?", s.tenantID)

	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where("sku LIKE ? OR product_name LIKE ?", search, search)
	}
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.LowStock {
		query = query.Where("quantity <= min_stock")
	}

	var total int64
	if err := query.Model(&models.InventoryRecord{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if filter.Limit == 0 {
		filter.Limit = 50
	}
	query = query.Limit(filter.Limit).Offset(filter.Offset).Order("sku ASC")

	var records []models.InventoryRecord
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}

	return &ListResult{
		Total:    total,
		Returned: len(records),
		Offset:   filter.Offset,
		Data:     records,
	}, nil
}

// GetBySKU retrieves a single inventory record by SKU
func (s *InventoryService) GetBySKU(ctx context.Context, sku string) (*models.InventoryRecord, error) {
	var record models.InventoryRecord
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND sku = ?", s.tenantID, sku).
		First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &record, err
}

// Update updates an inventory record
func (s *InventoryService) Update(ctx context.Context, sku string, updates map[string]interface{}) error {
	return s.db.WithContext(ctx).
		Model(&models.InventoryRecord{}).
		Where("tenant_id = ? AND sku = ?", s.tenantID, sku).
		Updates(updates).Error
}

// Delete deletes an inventory record
func (s *InventoryService) Delete(ctx context.Context, sku string) error {
	return s.db.WithContext(ctx).
		Where("tenant_id = ? AND sku = ?", s.tenantID, sku).
		Delete(&models.InventoryRecord{}).Error
}

// GetCategories retrieves distinct categories
func (s *InventoryService) GetCategories(ctx context.Context) ([]string, error) {
	var categories []string
	err := s.db.WithContext(ctx).
		Model(&models.InventoryRecord{}).
		Where("tenant_id = ? AND category != ''", s.tenantID).
		Distinct("category").
		Pluck("category", &categories).Error
	return categories, err
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
		data, _ := json.Marshal(input.AllColumns)
		existing.AllColumns = string(data)
	}
	if input.SelectedColumns != nil {
		data, _ := json.Marshal(input.SelectedColumns)
		existing.SelectedColumns = string(data)
	}

	return s.db.WithContext(ctx).Save(&existing).Error
}

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

// ============================================================================
// CRUD Operations by Key Value (for frontend inventory table editing)
// Added: 2026-01-27 - Fix 404 error on PUT /api/inventory/:keyValue
// ============================================================================

// GetByKeyValue retrieves inventory record by key value (e.g., SKU value)
// Returns nil if not found (not an error)
func (s *InventoryService) GetByKeyValue(ctx context.Context, keyValue string) (*models.InventoryRecord, error) {
	var record models.InventoryRecord
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", s.tenantID, keyValue).
		First(&record).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil // Not found is not an error
	}
	return &record, err
}

// UpdateByKeyValue updates inventory record by key value
// Returns updated record or error if not found
func (s *InventoryService) UpdateByKeyValue(ctx context.Context, keyValue string, data map[string]interface{}) (*models.InventoryRecord, error) {
	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Update record
	result := s.db.WithContext(ctx).
		Model(&models.InventoryRecord{}).
		Where("tenant_id = ? AND key_value = ?", s.tenantID, keyValue).
		Updates(map[string]interface{}{
			"data":       string(jsonData),
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// Return updated record
	return s.GetByKeyValue(ctx, keyValue)
}

// CreateRecord creates a new inventory record
func (s *InventoryService) CreateRecord(ctx context.Context, keyColumnName, keyValue string, data map[string]interface{}) (*models.InventoryRecord, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	record := &models.InventoryRecord{
		ID:            generateUUID(),
		TenantID:      s.tenantID,
		KeyValue:      keyValue,
		KeyColumnName: keyColumnName,
		Data:          string(jsonData),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = s.db.WithContext(ctx).Create(record).Error
	return record, err
}

// DeleteByKeyValue deletes inventory record by key value
// Returns error if record not found
func (s *InventoryService) DeleteByKeyValue(ctx context.Context, keyValue string) error {
	result := s.db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", s.tenantID, keyValue).
		Delete(&models.InventoryRecord{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
