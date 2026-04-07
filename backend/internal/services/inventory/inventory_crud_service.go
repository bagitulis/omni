package inventory

import (
	"context"
	"encoding/json"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

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
