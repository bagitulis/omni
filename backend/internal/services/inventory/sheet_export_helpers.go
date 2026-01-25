package inventory

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// exportToCSV exports inventory records to CSV format
// Uses JSONB data for dynamic columns
func exportToCSV(records []models.InventoryRecord) ([]byte, string, string, error) {
	if len(records) == 0 {
		return nil, "", "", fmt.Errorf("no records to export")
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Get all unique column names from first record
	firstData := GetDataMap(records[0])
	var headers []string
	for key := range firstData {
		headers = append(headers, key)
	}

	// Write header
	if err := writer.Write(headers); err != nil {
		return nil, "", "", err
	}

	// Write data
	for _, r := range records {
		data := GetDataMap(r)
		row := make([]string, len(headers))
		for i, header := range headers {
			if val, ok := data[header]; ok {
				row[i] = fmt.Sprintf("%v", val)
			}
		}
		if err := writer.Write(row); err != nil {
			return nil, "", "", err
		}
	}

	writer.Flush()
	filename := fmt.Sprintf("inventory_%s.csv", time.Now().Format("20060102_150405"))
	return buf.Bytes(), "text/csv", filename, nil
}

// buildSheetData builds data matrix from JSONB records
func buildSheetData(records []models.InventoryRecord, fields []string) [][]interface{} {
	if len(records) == 0 {
		return [][]interface{}{}
	}

	firstData := GetDataMap(records[0])
	if len(fields) == 0 {
		for key := range firstData {
			fields = append(fields, key)
		}
	}

	header := make([]interface{}, len(fields))
	for i, f := range fields {
		header[i] = f
	}

	data := [][]interface{}{header}

	for _, record := range records {
		recordData := GetDataMap(record)
		row := make([]interface{}, len(fields))
		for i, field := range fields {
			if val, ok := recordData[field]; ok {
				row[i] = val
			} else {
				row[i] = ""
			}
		}
		data = append(data, row)
	}

	return data
}

// parseRowToJSONBRecord parses a sheet row into JSONB data
func parseRowToJSONBRecord(headers []string, row []interface{}, keyColumn string) (map[string]interface{}, string) {
	data := make(map[string]interface{})
	var keyValue string

	for i, header := range headers {
		if i >= len(row) {
			break
		}
		value := fmt.Sprintf("%v", row[i])
		data[header] = value

		if header == keyColumn {
			keyValue = value
		}
	}

	return data, keyValue
}

// upsertJSONBRecord inserts or updates a JSONB inventory record
func upsertJSONBRecord(ctx context.Context, db *gorm.DB, tenantID string, data map[string]interface{}, keyValue, keyColumnName string, result *ImportResult) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	var existing models.InventoryRecord
	err = db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", tenantID, keyValue).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		record := models.InventoryRecord{
			ID:            uuid.NewString(),
			TenantID:      tenantID,
			Data:          string(jsonData),
			KeyValue:      keyValue,
			KeyColumnName: keyColumnName,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := db.WithContext(ctx).Create(&record).Error; err != nil {
			return err
		}
		result.RowsImported++
	} else if err == nil {
		existing.Data = string(jsonData)
		existing.UpdatedAt = time.Now()
		if err := db.WithContext(ctx).Save(&existing).Error; err != nil {
			return err
		}
		result.RowsUpdated++
	} else {
		return err
	}

	return nil
}

// buildSKUSet creates a lookup set for SKUs
func buildSKUSet(skus []string) map[string]bool {
	skuSet := make(map[string]bool)
	for _, sku := range skus {
		skuSet[sku] = true
	}
	return skuSet
}

// syncSingleJSONBRecord syncs a single JSONB inventory record
func syncSingleJSONBRecord(ctx context.Context, db *gorm.DB, tenantID string, data map[string]interface{}, keyValue, keyColumnName string, result *SyncResult) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	var existing models.InventoryRecord
	err = db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", tenantID, keyValue).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		record := models.InventoryRecord{
			ID:            uuid.NewString(),
			TenantID:      tenantID,
			Data:          string(jsonData),
			KeyValue:      keyValue,
			KeyColumnName: keyColumnName,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := db.WithContext(ctx).Create(&record).Error; err == nil {
			result.Added++
		}
	} else if err == nil {
		existing.Data = string(jsonData)
		existing.UpdatedAt = time.Now()
		if err := db.WithContext(ctx).Save(&existing).Error; err == nil {
			result.Updated++
		}
	}
}
