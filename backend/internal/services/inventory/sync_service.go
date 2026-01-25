package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// generateUUID creates a new UUID string
func generateUUID() string {
	return uuid.New().String()
}

// SyncService handles inventory synchronization
type SyncService struct {
	db           *gorm.DB
	tenantID     string
	sheetsClient SheetsClient
}

// NewSyncService creates a new sync service
func NewSyncService(db *gorm.DB, tenantID string, sheetsClient SheetsClient) *SyncService {
	return &SyncService{db: db, tenantID: tenantID, sheetsClient: sheetsClient}
}

// SheetSyncResult represents sync operation result from sheets
type SheetSyncResult struct {
	Status           string    `json:"status"` // SUCCESS, ERROR, PARTIAL
	Message          string    `json:"message"`
	TotalRecords     int       `json:"totalRecords"`
	NewRecords       int       `json:"newRecords"`
	UpdatedRecords   int       `json:"updatedRecords"`
	UnchangedRecords int       `json:"unchangedRecords"`
	FailedRecords    int       `json:"failedRecords"`
	Duration         int       `json:"durationMs"`
	HeadersChanged   bool      `json:"headersChanged"`
	Headers          []string  `json:"headers,omitempty"`
	SyncedAt         time.Time `json:"syncedAt"`
}

// SyncFromSheets syncs inventory from Google Sheets
func (s *SyncService) SyncFromSheets(ctx context.Context, spreadsheetID, sheetName string) (*SheetSyncResult, error) {
	startTime := time.Now()

	// Get settings
	invSvc := NewInventoryService(s.db, s.tenantID)
	settings, err := invSvc.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	// Use provided or default values
	if spreadsheetID == "" && settings != nil {
		spreadsheetID = settings.SpreadsheetID
	}
	if sheetName == "" && settings != nil {
		sheetName = settings.SheetName
	}

	if spreadsheetID == "" {
		return &SheetSyncResult{Status: "ERROR", Message: "Spreadsheet ID not configured"}, nil
	}

	// Read data from sheet
	sheetRange := sheetName
	if sheetRange == "" {
		sheetRange = "Sheet1"
	}

	data, err := s.sheetsClient.ReadRange(ctx, spreadsheetID, sheetRange)
	if err != nil {
		return &SheetSyncResult{Status: "ERROR", Message: err.Error()}, nil
	}

	if len(data) < 2 {
		return &SheetSyncResult{Status: "SUCCESS", Message: "No data rows found", TotalRecords: 0}, nil
	}

	// Parse headers
	headerRow := 0
	if settings != nil && settings.HeaderRow > 0 {
		headerRow = settings.HeaderRow - 1
	}
	headers := parseHeaders(data[headerRow])

	// Check if headers changed
	headersHash := hashStrings(headers)
	headersChanged := settings != nil && settings.LastHeadersHash != "" && settings.LastHeadersHash != headersHash

	// Parse data rows
	dataStartRow := headerRow + 1
	if settings != nil && settings.DataStartRow > 0 {
		dataStartRow = settings.DataStartRow - 1
	}

	keyColumn := "SKU"
	if settings != nil && settings.KeyColumn != "" {
		keyColumn = settings.KeyColumn
	}

	result := s.processRows(ctx, data[dataStartRow:], headers, keyColumn)
	result.Duration = int(time.Since(startTime).Milliseconds())
	result.HeadersChanged = headersChanged
	result.Headers = headers
	result.SyncedAt = time.Now()

	// Update settings with new hash and columns
	if settings != nil {
		settings.LastHeadersHash = headersHash
		settings.LastSyncTimestamp = &result.SyncedAt
		settings.LastSyncStatus = result.Status
		
		// Always update AllColumns with discovered headers from sheet
		allColumnsJSON, _ := json.Marshal(headers)
		settings.AllColumns = string(allColumnsJSON)
		
		// Auto-populate SelectedColumns if empty (first sync scenario)
		if settings.SelectedColumns == "" || settings.SelectedColumns == "[]" || settings.SelectedColumns == "null" {
			settings.SelectedColumns = string(allColumnsJSON)
		}
		
		_ = invSvc.UpdateSettings(ctx, settings)
	}

	// Record sync history
	history := &models.InventorySyncHistory{
		Status:           result.Status,
		TotalRecords:     result.TotalRecords,
		NewRecords:       result.NewRecords,
		UpdatedRecords:   result.UpdatedRecords,
		UnchangedRecords: result.UnchangedRecords,
		FailedRecords:    result.FailedRecords,
		Duration:         result.Duration,
		HeadersChanged:   result.HeadersChanged,
	}
	_ = invSvc.RecordSyncHistory(ctx, history)

	return result, nil
}

func (s *SyncService) processRows(ctx context.Context, rows [][]interface{}, headers []string, keyColumn string) *SheetSyncResult {
	result := &SheetSyncResult{Status: "SUCCESS"}
	keyIdx := findColumnIndex(headers, keyColumn)
	if keyIdx < 0 {
		result.Status = "ERROR"
		result.Message = fmt.Sprintf("Key column '%s' not found", keyColumn)
		return result
	}

	for _, row := range rows {
		if len(row) <= keyIdx || row[keyIdx] == nil {
			result.FailedRecords++
			continue
		}

		keyValue := fmt.Sprintf("%v", row[keyIdx])
		if keyValue == "" {
			result.FailedRecords++
			continue
		}

		// Build record data as JSON (match PostgreSQL JSONB format)
		recordData := make(map[string]interface{})
		for i, h := range headers {
			if i < len(row) && row[i] != nil {
				recordData[h] = row[i]
			} else {
				recordData[h] = ""
			}
		}

		dataJSON, err := json.Marshal(recordData)
		if err != nil {
			result.FailedRecords++
			continue
		}

		record := &models.InventoryRecord{
			TenantID:      s.tenantID,
			Data:          string(dataJSON),
			KeyValue:      keyValue,
			KeyColumnName: keyColumn,
			UpdatedAt:     time.Now(),
		}

		status, err := s.upsertRecord(ctx, record)
		if err != nil {
			result.FailedRecords++
			continue
		}

		result.TotalRecords++
		switch status {
		case "new":
			result.NewRecords++
		case "updated":
			result.UpdatedRecords++
		case "unchanged":
			result.UnchangedRecords++
		}
	}

	if result.FailedRecords > 0 && result.NewRecords+result.UpdatedRecords > 0 {
		result.Status = "PARTIAL"
	} else if result.FailedRecords > 0 {
		result.Status = "ERROR"
	}

	return result
}

func (s *SyncService) upsertRecord(ctx context.Context, record *models.InventoryRecord) (string, error) {
	var existing models.InventoryRecord
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND key_column_name = ? AND key_value = ?", 
			s.tenantID, record.KeyColumnName, record.KeyValue).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Generate UUID for new record
		record.ID = generateUUID()
		record.CreatedAt = time.Now()
		if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
			return "", err
		}
		return "new", nil
	}
	if err != nil {
		return "", err
	}

	// Check if data changed
	if existing.Data == record.Data {
		return "unchanged", nil
	}

	// Update existing record
	existing.Data = record.Data
	existing.UpdatedAt = time.Now()
	if err := s.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return "", err
	}
	return "updated", nil
}

