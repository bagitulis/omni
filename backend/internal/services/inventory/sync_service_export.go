package inventory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog/log"
)

// SyncToSheets syncs inventory data FROM database TO Google Sheets
func (s *SyncService) SyncToSheets(ctx context.Context, spreadsheetID, sheetName string, lockedColumns []string) (*SheetSyncResult, error) {
	startTime := time.Now()

	// Get settings
	invSvc := NewInventoryService(s.db, s.tenantID)
	settings, err := invSvc.GetSettings(ctx)
	if err != nil && spreadsheetID == "" {
		return &SheetSyncResult{Status: "ERROR", Message: "Settings not found"}, nil
	}

	// Use provided or default values
	if spreadsheetID == "" && settings != nil {
		spreadsheetID = settings.SpreadsheetID
	}
	if sheetName == "" && settings != nil {
		sheetName = settings.SheetName
	}

	if spreadsheetID == "" || sheetName == "" {
		return &SheetSyncResult{Status: "ERROR", Message: "Spreadsheet not configured"}, nil
	}

	// Get all records from database (capped at 50000 to prevent OOM)
	var records []models.InventoryRecord
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", s.tenantID).Limit(50000).Find(&records).Error; err != nil {
		return &SheetSyncResult{Status: "ERROR", Message: err.Error()}, nil
	}

	if len(records) == 0 {
		return &SheetSyncResult{Status: "SUCCESS", Message: "No records to export"}, nil
	}

	// Get selected columns from settings
	var selectedColumns []string
	if settings != nil && settings.SelectedColumns != "" {
		json.Unmarshal([]byte(settings.SelectedColumns), &selectedColumns)
	}

	// Build headers from first record if no selected columns
	if len(selectedColumns) == 0 {
		var firstData map[string]interface{}
		if err := json.Unmarshal([]byte(records[0].Data), &firstData); err == nil {
			for k := range firstData {
				selectedColumns = append(selectedColumns, k)
			}
		}
	}

	// Filter out locked columns
	lockedSet := make(map[string]bool)
	for _, col := range lockedColumns {
		lockedSet[col] = true
	}

	// Prepare data for export
	var values [][]interface{}

	// Add header row
	headers := make([]interface{}, 0, len(selectedColumns))
	for _, col := range selectedColumns {
		if !lockedSet[col] {
			headers = append(headers, col)
		}
	}
	values = append(values, headers)

	// Add data rows
	for _, record := range records {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(record.Data), &data); err != nil {
			continue
		}

		row := make([]interface{}, 0, len(selectedColumns))
		for _, col := range selectedColumns {
			if !lockedSet[col] {
				if val, ok := data[col]; ok {
					row = append(row, val)
				} else {
					row = append(row, "")
				}
			}
		}
		values = append(values, row)
	}

	// Clear existing data and write new data
	if err := s.sheetsClient.ClearRange(ctx, spreadsheetID, sheetName); err != nil {
		return &SheetSyncResult{Status: "ERROR", Message: "Failed to clear sheet: " + err.Error()}, nil
	}

	sheetRange := fmt.Sprintf("%s!A1", sheetName)
	if err := s.sheetsClient.WriteRange(ctx, spreadsheetID, sheetRange, values); err != nil {
		return &SheetSyncResult{Status: "ERROR", Message: "Failed to write data: " + err.Error()}, nil
	}

	result := &SheetSyncResult{
		Status:           "SUCCESS",
		Message:          fmt.Sprintf("Exported %d records to Google Sheets", len(records)),
		TotalRecords:     len(records),
		UpdatedRecords:   len(records),
		UnchangedRecords: 0,
		Duration:         int(time.Since(startTime).Milliseconds()),
		SyncedAt:         time.Now(),
	}

	// Update sync status
	if settings != nil {
		settings.LastSyncTimestamp = &result.SyncedAt
		settings.LastSyncStatus = result.Status
		if updateErr := invSvc.UpdateSettings(ctx, settings); updateErr != nil {
			log.Warn().Err(updateErr).Str("tenant_id", s.tenantID).Msg("Failed to update inventory settings after export")
		}
	}

	// Record sync history
	history := &models.InventorySyncHistory{
		Status:         result.Status,
		TotalRecords:   result.TotalRecords,
		UpdatedRecords: result.UpdatedRecords,
		Duration:       result.Duration,
	}
	if historyErr := invSvc.RecordSyncHistory(ctx, history); historyErr != nil {
		log.Warn().Err(historyErr).Str("tenant_id", s.tenantID).Msg("Failed to record export sync history")
	}

	return result, nil
}
