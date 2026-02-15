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
	Search            string
	SyncStatus        []string // "synced", "not_synced", "error"
	StockStatus       string   // "in_stock", "low_stock", "out_of_stock", ""
	Platform          []string // "shopee", "lazada", "tiktok"
	LowStockThreshold int      // For stock_status filtering
	Limit             int
	Offset            int
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
		query = query.Where("key_value LIKE ?", search)
	}

	query = buildSyncStatusFilter(query, filter)
	query = buildPlatformFilter(query, filter, s.tenantID)
	query = buildStockStatusFilter(query, filter)

	var total int64
	if err := query.Model(&models.InventoryRecord{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if filter.Limit == 0 {
		filter.Limit = 50
	}
	query = query.Limit(filter.Limit).Offset(filter.Offset).Order("key_value ASC")

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

// PlatformStatusItem represents platform sync status for a SKU
type PlatformStatusItem struct {
	Platform          string  `json:"platform"`
	PlatformProductID string  `json:"platform_product_id"`
	PlatformItemID    string  `json:"platform_item_id"`
	PlatformSKU       string  `json:"platform_sku"`
	Status            string  `json:"status"`
	Stock             int     `json:"stock"`
	Price             float64 `json:"price"`
	LastSyncAt        string  `json:"last_sync_at"`
	ErrorMessage      string  `json:"error_message,omitempty"`
}

// GetRecordsWithPlatformStatus retrieves inventory records with platform status enrichment
func (s *InventoryService) GetRecordsWithPlatformStatus(ctx context.Context, filter ListFilter) (*ListResult, map[string][]PlatformStatusItem, error) {
	// Get base records
	result, err := s.GetRecords(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	// Extract SKUs
	skus := make([]string, 0, len(result.Data))
	for _, rec := range result.Data {
		skus = append(skus, rec.KeyValue)
	}

	// Fetch platform statuses
	platformMap := make(map[string][]PlatformStatusItem)
	if len(skus) > 0 {
		var platformStatuses []models.InventorySkuPlatformStatus
		err = s.db.WithContext(ctx).
			Where("tenant_id = ? AND sku IN ?", s.tenantID, skus).
			Find(&platformStatuses).Error
		if err != nil {
			return nil, nil, err
		}

		// Group by SKU
		for _, ps := range platformStatuses {
			platformMap[ps.SKU] = append(platformMap[ps.SKU], PlatformStatusItem{
				Platform:          ps.Platform,
				PlatformProductID: ps.PlatformProductID,
				PlatformItemID:    ps.PlatformItemID,
				PlatformSKU:       ps.PlatformSKU,
				Status:            ps.Status,
				Stock:             ps.Stock,
				Price:             ps.Price,
				LastSyncAt:        ps.LastCheckedAt.Format(time.RFC3339),
			})
		}
	}

	return result, platformMap, nil
}

// GetBySKU retrieves a single inventory record by SKU
func (s *InventoryService) GetBySKU(ctx context.Context, sku string) (*models.InventoryRecord, error) {
	var record models.InventoryRecord
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", s.tenantID, sku).
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
		Where("tenant_id = ? AND key_value = ?", s.tenantID, sku).
		Updates(updates).Error
}

// Delete deletes an inventory record
func (s *InventoryService) Delete(ctx context.Context, sku string) error {
	return s.db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", s.tenantID, sku).
		Delete(&models.InventoryRecord{}).Error
}

// GetCategories retrieves distinct categories
func (s *InventoryService) GetCategories(ctx context.Context) ([]string, error) {
	// Categories are stored in JSONB data field, not as a direct column
	return []string{}, nil
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
